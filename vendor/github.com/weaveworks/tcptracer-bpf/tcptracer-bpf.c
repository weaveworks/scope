#include <linux/kconfig.h>

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wgnu-variable-sized-type-not-at-end"
#include <linux/ptrace.h>
#pragma clang diagnostic pop
#include <linux/version.h>
#include <linux/bpf.h>
#include "bpf_helpers.h"
#include "tcptracer-bpf.h"

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wtautological-compare"
#include <net/sock.h>
#pragma clang diagnostic pop
#include <net/inet_sock.h>
#include <net/net_namespace.h>

/* This is a key/value store with the keys being the cpu number
 * and the values being a perf file descriptor.
 */
struct bpf_map_def SEC("maps/tcp_event_ipv4") tcp_event_ipv4 = {
	.type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
	.key_size = sizeof(int),
	.value_size = sizeof(__u32),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being the cpu number
 * and the values being a perf file descriptor.
 */
struct bpf_map_def SEC("maps/tcp_event_ipv6") tcp_event_ipv6 = {
	.type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
	.key_size = sizeof(int),
	.value_size = sizeof(__u32),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* These maps are used to match the kprobe & kretprobe of connect */

/* This is a key/value store with the keys being a pid
 * and the values being a struct sock *.
 */
struct bpf_map_def SEC("maps/connectsock_ipv4") connectsock_ipv4 = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(void *),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being a pid
 * and the values being a struct sock *.
 */
struct bpf_map_def SEC("maps/connectsock_ipv6") connectsock_ipv6 = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(void *),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being an ipv4_tuple_t
 * and the values being a struct pid_comm_t.
 */
struct bpf_map_def SEC("maps/tuplepid_ipv4") tuplepid_ipv4 = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(struct ipv4_tuple_t),
	.value_size = sizeof(struct pid_comm_t),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being an ipv6_tuple_t
 * and the values being a struct pid_comm_t.
 */
struct bpf_map_def SEC("maps/tuplepid_ipv6") tuplepid_ipv6 = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(struct ipv6_tuple_t),
	.value_size = sizeof(struct pid_comm_t),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being a pid
 * and the values being a fd unsigned int.
 */
struct bpf_map_def SEC("maps/fdinstall_ret") fdinstall_ret = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(unsigned int),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* This is a key/value store with the keys being a pid (tgid)
 * and the values being a boolean.
 */
struct bpf_map_def SEC("maps/fdinstall_pids") fdinstall_pids = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u32),
	.value_size = sizeof(__u32),
	.max_entries = 1024,
	.pinning = 0,
	.namespace = "",
};

/* http://stackoverflow.com/questions/1001307/detecting-endianness-programmatically-in-a-c-program */
__attribute__((always_inline))
static bool is_big_endian(void)
{
	union {
		uint32_t i;
		char c[4];
	} bint = {0x01020304};

	return bint.c[0] == 1;
}

/* check if IPs are IPv4 mapped to IPv6 ::ffff:xxxx:xxxx
 * https://tools.ietf.org/html/rfc4291#section-2.5.5
 * the addresses are stored in network byte order so IPv4 adddress is stored
 * in the most significant 32 bits of part saddr_l and daddr_l.
 * Meanwhile the end of the mask is stored in the least significant 32 bits.
 */
__attribute__((always_inline))
static bool is_ipv4_mapped_ipv6(u64 saddr_h, u64 saddr_l, u64 daddr_h, u64 daddr_l) {
	if (is_big_endian()) {
		return ((saddr_h == 0 && ((u32)(saddr_l >> 32) == 0x0000FFFF)) ||
                        (daddr_h == 0 && ((u32)(daddr_l >> 32) == 0x0000FFFF)));
	} else {
		return ((saddr_h == 0 && ((u32)saddr_l == 0xFFFF0000)) ||
                        (daddr_h == 0 && ((u32)daddr_l == 0xFFFF0000)));
	}
}

struct bpf_map_def SEC("maps/tcptracer_status") tcptracer_status = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(struct tcptracer_status_t),
	.max_entries = 1,
	.pinning = 0,
	.namespace = "",
};

__attribute__((always_inline))
static int are_offsets_ready_v4(struct tcptracer_status_t *status, struct sock *skp, u64 pid) {
	u64 zero = 0;

	switch (status->state) {
		case TCPTRACER_STATE_UNINITIALIZED:
			return 0;
		case TCPTRACER_STATE_CHECKING:
			break;
		case TCPTRACER_STATE_CHECKED:
			return 0;
		case TCPTRACER_STATE_READY:
			return 1;
		default:
			return 0;
	}

	// Only accept the exact pid & tid. Extraneous connections from other
	// threads must be ignored here. Userland must take care to generate
	// connections from the correct thread. In Golang, this can be achieved
	// with runtime.LockOSThread.
	if (status->pid_tgid != pid)
		return 0;

	struct tcptracer_status_t new_status = { };
	new_status.state = TCPTRACER_STATE_CHECKED;
	new_status.pid_tgid = status->pid_tgid;
	new_status.what = status->what;
	new_status.offset_saddr = status->offset_saddr;
	new_status.offset_daddr = status->offset_daddr;
	new_status.offset_sport = status->offset_sport;
	new_status.offset_dport = status->offset_dport;
	new_status.offset_netns = status->offset_netns;
	new_status.offset_ino = status->offset_ino;
	new_status.offset_family = status->offset_family;
	new_status.offset_daddr_ipv6 = status->offset_daddr_ipv6;
	new_status.err = 0;
	new_status.saddr = status->saddr;
	new_status.daddr = status->daddr;
	new_status.sport = status->sport;
	new_status.dport = status->dport;
	new_status.netns = status->netns;
	new_status.family = status->family;

	int i;
	for (i = 0; i < 4; i++) {
		new_status.daddr_ipv6[i] = status->daddr_ipv6[i];
	}

	u32 possible_saddr;
	u32 possible_daddr;
	u16 possible_sport;
	u16 possible_dport;
	possible_net_t *possible_skc_net;
	u32 possible_netns;
	u16 possible_family;
	long ret = 0;

	switch (status->what) {
		case GUESS_SADDR:
			possible_saddr = 0;
			bpf_probe_read(&possible_saddr, sizeof(possible_saddr), ((char *)skp) + status->offset_saddr);
			new_status.saddr = possible_saddr;
			break;
		case GUESS_DADDR:
			possible_daddr = 0;
			bpf_probe_read(&possible_daddr, sizeof(possible_daddr), ((char *)skp) + status->offset_daddr);
			new_status.daddr = possible_daddr;
			break;
		case GUESS_FAMILY:
			possible_family = 0;
			bpf_probe_read(&possible_family, sizeof(possible_family), ((char *)skp) + status->offset_family);
			new_status.family = possible_family;
			break;
		case GUESS_SPORT:
			possible_sport = 0;
			bpf_probe_read(&possible_sport, sizeof(possible_sport), ((char *)skp) + status->offset_sport);
			new_status.sport = possible_sport;
			break;
		case GUESS_DPORT:
			possible_dport = 0;
			bpf_probe_read(&possible_dport, sizeof(possible_dport), ((char *)skp) + status->offset_dport);
			new_status.dport = possible_dport;
			break;
		case GUESS_NETNS:
			possible_netns = 0;
			possible_skc_net = NULL;
			bpf_probe_read(&possible_skc_net, sizeof(possible_net_t *), ((char *)skp) + status->offset_netns);
			// if we get a kernel fault, it means possible_skc_net
			// is an invalid pointer, signal an error so we can go
			// to the next offset_netns
			ret = bpf_probe_read(&possible_netns, sizeof(possible_netns), ((char *)possible_skc_net) + status->offset_ino);
			if (ret == -EFAULT) {
				new_status.err = 1;
				break;
			}
			new_status.netns = possible_netns;
			break;
		default:
			// not for us
			return 0;
	}

	bpf_map_update_elem(&tcptracer_status, &zero, &new_status, BPF_ANY);

	return 0;
}

__attribute__((always_inline))
static int are_offsets_ready_v6(struct tcptracer_status_t *status, struct sock *skp, u64 pid) {
	u64 zero = 0;

	switch (status->state) {
		case TCPTRACER_STATE_UNINITIALIZED:
			return 0;
		case TCPTRACER_STATE_CHECKING:
			break;
		case TCPTRACER_STATE_CHECKED:
			return 0;
		case TCPTRACER_STATE_READY:
			return 1;
		default:
			return 0;
	}

	// Only accept the exact pid & tid. Extraneous connections from other
	// threads must be ignored here. Userland must take care to generate
	// connections from the correct thread. In Golang, this can be achieved
	// with runtime.LockOSThread.
	if (status->pid_tgid != pid)
		return 0;

	struct tcptracer_status_t new_status = { };
	new_status.state = TCPTRACER_STATE_CHECKED;
	new_status.pid_tgid = status->pid_tgid;
	new_status.what = status->what;
	new_status.offset_saddr = status->offset_saddr;
	new_status.offset_daddr = status->offset_daddr;
	new_status.offset_sport = status->offset_sport;
	new_status.offset_dport = status->offset_dport;
	new_status.offset_netns = status->offset_netns;
	new_status.offset_ino = status->offset_ino;
	new_status.offset_family = status->offset_family;
	new_status.offset_daddr_ipv6 = status->offset_daddr_ipv6;
	new_status.err = 0;
	new_status.saddr = status->saddr;
	new_status.daddr = status->daddr;
	new_status.sport = status->sport;
	new_status.dport = status->dport;
	new_status.netns = status->netns;
	new_status.family = status->family;

	int i;
	for (i = 0; i < 4; i++) {
		new_status.daddr_ipv6[i] = status->daddr_ipv6[i];
	}

	u32 possible_daddr_ipv6[4] = { };
	switch (status->what) {
		case GUESS_DADDR_IPV6:
			bpf_probe_read(&possible_daddr_ipv6, sizeof(possible_daddr_ipv6), ((char *)skp) + status->offset_daddr_ipv6);

			int i;
			for (i = 0; i < 4; i++) {
				new_status.daddr_ipv6[i] = possible_daddr_ipv6[i];
			}
			break;
		default:
			// not for us
			return 0;
	}

	bpf_map_update_elem(&tcptracer_status, &zero, &new_status, BPF_ANY);

	return 0;
}

__attribute__((always_inline))
static bool check_family(struct sock *sk, u16 expected_family) {
	struct tcptracer_status_t *status;
	u64 zero = 0;
	u16 family;
	family = 0;

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state != TCPTRACER_STATE_READY) {
		return 0;
	}

	bpf_probe_read(&family, sizeof(u16), ((char *)sk) + status->offset_family);

	return family == expected_family;
}

__attribute__((always_inline))
static int read_ipv4_tuple(struct ipv4_tuple_t *tuple, struct tcptracer_status_t *status, struct sock *skp)
{
	u32 saddr, daddr, net_ns_inum;
	u16 sport, dport;
	possible_net_t *skc_net;

	saddr = 0;
	daddr = 0;
	sport = 0;
	dport = 0;
	skc_net = NULL;
	net_ns_inum = 0;

	bpf_probe_read(&saddr, sizeof(saddr), ((char *)skp) + status->offset_saddr);
	bpf_probe_read(&daddr, sizeof(daddr), ((char *)skp) + status->offset_daddr);
	bpf_probe_read(&sport, sizeof(sport), ((char *)skp) + status->offset_sport);
	bpf_probe_read(&dport, sizeof(dport), ((char *)skp) + status->offset_dport);
	// Get network namespace id
	bpf_probe_read(&skc_net, sizeof(void *), ((char *)skp) + status->offset_netns);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), ((char *)skc_net) + status->offset_ino);

	tuple->saddr = saddr;
	tuple->daddr = daddr;
	tuple->sport = sport;
	tuple->dport = dport;
	tuple->netns = net_ns_inum;

	// if addresses or ports are 0, ignore
	if (saddr == 0 || daddr == 0 || sport == 0 || dport == 0) {
		return 0;
	}

	return 1;
}

__attribute__((always_inline))
static int read_ipv6_tuple(struct ipv6_tuple_t *tuple, struct tcptracer_status_t *status, struct sock *skp)
{
	u32 net_ns_inum;
	u16 sport, dport;
	u64 saddr_h, saddr_l, daddr_h, daddr_l;
	possible_net_t *skc_net;

	saddr_h = 0;
	saddr_l = 0;
	daddr_h = 0;
	daddr_l = 0;
	sport = 0;
	dport = 0;
	skc_net = NULL;
	net_ns_inum = 0;

	bpf_probe_read(&saddr_h, sizeof(saddr_h), ((char *)skp) + status->offset_daddr_ipv6 + 2 * sizeof(u64));
	bpf_probe_read(&saddr_l, sizeof(saddr_l), ((char *)skp) + status->offset_daddr_ipv6 + 3 * sizeof(u64));
	bpf_probe_read(&daddr_h, sizeof(daddr_h), ((char *)skp) + status->offset_daddr_ipv6);
	bpf_probe_read(&daddr_l, sizeof(daddr_l), ((char *)skp) + status->offset_daddr_ipv6 + sizeof(u64));
	bpf_probe_read(&sport, sizeof(sport), ((char *)skp) + status->offset_sport);
	bpf_probe_read(&dport, sizeof(dport), ((char *)skp) + status->offset_dport);
	// Get network namespace id
	bpf_probe_read(&skc_net, sizeof(void *), ((char *)skp) + status->offset_netns);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), ((char *)skc_net) + status->offset_ino);

	tuple->saddr_h = saddr_h;
	tuple->saddr_l = saddr_l;
	tuple->daddr_h = daddr_h;
	tuple->daddr_l = daddr_l;
	tuple->sport = sport;
	tuple->dport = dport;
	tuple->netns = net_ns_inum;

	// if addresses or ports are 0, ignore
	if (!(saddr_h || saddr_l) || !(daddr_h || daddr_l) || sport == 0 || dport == 0) {
		return 0;
	}

	return 1;
}

SEC("kprobe/tcp_v4_connect")
int kprobe__tcp_v4_connect(struct pt_regs *ctx)
{
	struct sock *sk;
	u64 pid = bpf_get_current_pid_tgid();

	sk = (struct sock *) PT_REGS_PARM1(ctx);

	bpf_map_update_elem(&connectsock_ipv4, &pid, &sk, BPF_ANY);

	return 0;
}

SEC("kretprobe/tcp_v4_connect")
int kretprobe__tcp_v4_connect(struct pt_regs *ctx)
{
	int ret = PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();
	struct sock **skpp;
	u64 zero = 0;
	struct tcptracer_status_t *status;

	skpp = bpf_map_lookup_elem(&connectsock_ipv4, &pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	struct sock *skp = *skpp;

	bpf_map_delete_elem(&connectsock_ipv4, &pid);

	if (ret != 0) {
		// failed to send SYNC packet, may not have populated
		// socket __sk_common.{skc_rcv_saddr, ...}
		return 0;
	}

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state == TCPTRACER_STATE_UNINITIALIZED) {
		return 0;
	}

	if (!are_offsets_ready_v4(status, skp, pid)) {
		return 0;
	}

	// output
	struct ipv4_tuple_t t = { };
	if (!read_ipv4_tuple(&t, status, skp)) {
		return 0;
	}

	struct pid_comm_t p = { .pid = pid };
	bpf_get_current_comm(p.comm, sizeof(p.comm));
	bpf_map_update_elem(&tuplepid_ipv4, &t, &p, BPF_ANY);

	return 0;
}

SEC("kprobe/tcp_v6_connect")
int kprobe__tcp_v6_connect(struct pt_regs *ctx)
{
	struct sock *sk;
	u64 pid = bpf_get_current_pid_tgid();

	sk = (struct sock *) PT_REGS_PARM1(ctx);

	bpf_map_update_elem(&connectsock_ipv6, &pid, &sk, BPF_ANY);

	return 0;
}

SEC("kretprobe/tcp_v6_connect")
int kretprobe__tcp_v6_connect(struct pt_regs *ctx)
{
	int ret = PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();
	u64 zero = 0;
	struct sock **skpp;
	struct tcptracer_status_t *status;
	skpp = bpf_map_lookup_elem(&connectsock_ipv6, &pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	bpf_map_delete_elem(&connectsock_ipv6, &pid);

	struct sock *skp = *skpp;

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state == TCPTRACER_STATE_UNINITIALIZED) {
		return 0;
	}

	if (!are_offsets_ready_v6(status, skp, pid)) {
		return 0;
	}

	if (ret != 0) {
		// failed to send SYNC packet, may not have populated
		// socket __sk_common.{skc_rcv_saddr, ...}
		return 0;
	}

	// output
	struct ipv6_tuple_t t = { };
	if (!read_ipv6_tuple(&t, status, skp)) {
		return 0;
	}

	struct pid_comm_t p = { };
	p.pid = pid;
	bpf_get_current_comm(p.comm, sizeof(p.comm));

	if (is_ipv4_mapped_ipv6(t.saddr_h, t.saddr_l, t.daddr_h, t.daddr_l)) {
		struct ipv4_tuple_t t4 = {
			.netns = t.netns,
			.saddr = (u32)(t.saddr_l >> 32),
			.daddr = (u32)(t.daddr_l >> 32),
			.sport = ntohs(t.sport),
			.dport = ntohs(t.dport),
		};
		bpf_map_update_elem(&tuplepid_ipv4, &t4, &p, BPF_ANY);
		return 0;
	}

	bpf_map_update_elem(&tuplepid_ipv6, &t, &p, BPF_ANY);
	return 0;
}

SEC("kprobe/tcp_set_state")
int kprobe__tcp_set_state(struct pt_regs *ctx)
{
	u32 cpu = bpf_get_smp_processor_id();
	struct sock *skp;
	struct tcptracer_status_t *status;
	int state;
	u64 zero = 0;
	skp =  (struct sock *) PT_REGS_PARM1(ctx);
	state = (int) PT_REGS_PARM2(ctx);

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state != TCPTRACER_STATE_READY) {
		return 0;
	}

	if (state != TCP_ESTABLISHED && state != TCP_CLOSE) {
		return 0;
	}

	if (check_family(skp, AF_INET)) {
		// output
		struct ipv4_tuple_t t = { };
		if (!read_ipv4_tuple(&t, status, skp)) {
			return 0;
		}
		if (state == TCP_CLOSE) {
			bpf_map_delete_elem(&tuplepid_ipv4, &t);
			return 0;
		}

		struct pid_comm_t *pp;

		pp = bpf_map_lookup_elem(&tuplepid_ipv4, &t);
		if (pp == 0) {
			return 0;	// missed entry
		}
		struct pid_comm_t p = { };
		bpf_probe_read(&p, sizeof(struct pid_comm_t), pp);

		struct tcp_ipv4_event_t evt4 = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_CONNECT,
			.pid = p.pid >> 32,
			.saddr = t.saddr,
			.daddr = t.daddr,
			.sport = ntohs(t.sport),
			.dport = ntohs(t.dport),
			.netns = t.netns,
		};
		int i;
		for (i = 0; i < TASK_COMM_LEN; i++) {
			evt4.comm[i] = p.comm[i];
		}

		bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt4, sizeof(evt4));
		bpf_map_delete_elem(&tuplepid_ipv4, &t);
	} else if (check_family(skp, AF_INET6)) {
		// output
		struct ipv6_tuple_t t = { };
		if (!read_ipv6_tuple(&t, status, skp)) {
			return 0;
		}
		if (state == TCP_CLOSE) {
			bpf_map_delete_elem(&tuplepid_ipv6, &t);
			return 0;
		}

		struct pid_comm_t *pp;
		pp = bpf_map_lookup_elem(&tuplepid_ipv6, &t);
		if (pp == 0) {
			return 0;       // missed entry
		}
		struct pid_comm_t p = { };
		bpf_probe_read(&p, sizeof(struct pid_comm_t), pp);
		struct tcp_ipv6_event_t evt6 = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_CONNECT,
			.pid = p.pid >> 32,
			.saddr_h = t.saddr_h,
			.saddr_l = t.saddr_l,
			.daddr_h = t.daddr_h,
			.daddr_l = t.daddr_l,
			.sport = ntohs(t.sport),
			.dport = ntohs(t.dport),
			.netns = t.netns,
		};
		int i;
		for (i = 0; i < TASK_COMM_LEN; i++) {
			evt6.comm[i] = p.comm[i];
		}

		bpf_perf_event_output(ctx, &tcp_event_ipv6, cpu, &evt6, sizeof(evt6));
		bpf_map_delete_elem(&tuplepid_ipv6, &t);
	}

	return 0;
}

SEC("kprobe/tcp_close")
int kprobe__tcp_close(struct pt_regs *ctx)
{
	struct sock *sk;
	struct tcptracer_status_t *status;
	u64 zero = 0;
	u64 pid = bpf_get_current_pid_tgid();
	u32 cpu = bpf_get_smp_processor_id();
	sk = (struct sock *) PT_REGS_PARM1(ctx);

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state != TCPTRACER_STATE_READY) {
		return 0;
	}

	u32 net_ns_inum;
	u16 sport, dport;
	sport = 0;
	dport = 0;

	// Get network namespace id
	possible_net_t *skc_net;

	skc_net = NULL;
	net_ns_inum = 0;
	bpf_probe_read(&skc_net, sizeof(possible_net_t *), ((char *)sk) + status->offset_netns);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), ((char *)skc_net) + status->offset_ino);

	if (check_family(sk, AF_INET)) {
		// output
		struct ipv4_tuple_t t = { };
		if (!read_ipv4_tuple(&t, status, sk)) {
			bpf_map_delete_elem(&tuplepid_ipv4, &t);
			return 0;
		}

		// output
		struct tcp_ipv4_event_t evt = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_CLOSE,
			.pid = pid >> 32,
			.saddr = t.saddr,
			.daddr = t.daddr,
			.sport = ntohs(t.sport),
			.dport = ntohs(t.dport),
			.netns = t.netns,
		};
		bpf_get_current_comm(&evt.comm, sizeof(evt.comm));

		bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt, sizeof(evt));
	} else if (check_family(sk, AF_INET6)) {
		// output
		struct ipv6_tuple_t t = { };
		if (!read_ipv6_tuple(&t, status, sk)) {
			bpf_map_delete_elem(&tuplepid_ipv6, &t);
			return 0;
		}

		if (is_ipv4_mapped_ipv6(t.saddr_h, t.saddr_l, t.daddr_h, t.daddr_l)) {
			struct tcp_ipv4_event_t evt4 = {
				.timestamp = bpf_ktime_get_ns(),
				.cpu = cpu,
				.type = TCP_EVENT_TYPE_CLOSE,
				.pid = pid >> 32,
				.saddr = (u32)(t.saddr_l >> 32),
				.daddr = (u32)(t.daddr_l >> 32),
				.sport = ntohs(t.sport),
				.dport = ntohs(t.dport),
				.netns = t.netns,
			};
			bpf_get_current_comm(&evt4.comm, sizeof(evt4.comm));
			if (evt4.saddr != 0 && evt4.daddr != 0 && evt4.sport != 0 && evt4.dport != 0) {
				bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt4, sizeof(evt4));
			}

			struct ipv4_tuple_t t = {
				t.saddr = evt4.saddr,
				t.daddr = evt4.daddr,
				t.sport = ntohs(evt4.sport),
				t.dport = ntohs(evt4.dport),
				t.netns = evt4.netns,
			};
			bpf_map_delete_elem(&tuplepid_ipv4, &t);
			return 0;
		}

		struct tcp_ipv6_event_t evt = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_CLOSE,
			.pid = pid >> 32,
			.saddr_h = t.saddr_h,
			.saddr_l = t.saddr_l,
			.daddr_h = t.daddr_h,
			.daddr_l = t.daddr_l,
			.sport = ntohs(t.sport),
			.dport = ntohs(t.dport),
			.netns = t.netns,
		};
		bpf_get_current_comm(&evt.comm, sizeof(evt.comm));

		bpf_perf_event_output(ctx, &tcp_event_ipv6, cpu, &evt, sizeof(evt));
	}
	return 0;
}

SEC("kretprobe/inet_csk_accept")
int kretprobe__inet_csk_accept(struct pt_regs *ctx)
{
	struct tcptracer_status_t *status;
	u64 zero = 0;
	struct sock *newsk = (struct sock *)PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();
	u32 cpu = bpf_get_smp_processor_id();

	if (newsk == NULL)
		return 0;

	status = bpf_map_lookup_elem(&tcptracer_status, &zero);
	if (status == NULL || status->state != TCPTRACER_STATE_READY) {
		return 0;
	}

	// pull in details
	u16 lport, dport;
	u32 net_ns_inum;

	lport = 0;
	dport = 0;

	bpf_probe_read(&dport, sizeof(dport), ((char *)newsk) + status->offset_dport);
        // lport is right after dport
	bpf_probe_read(&lport, sizeof(lport), ((char *)newsk) + status->offset_dport + sizeof(dport));
	// Get network namespace id
	possible_net_t *skc_net;

	skc_net = NULL;
	net_ns_inum = 0;
	bpf_probe_read(&skc_net, sizeof(possible_net_t *), ((char *)newsk) + status->offset_netns);
	bpf_probe_read(&net_ns_inum, sizeof(net_ns_inum), ((char *)skc_net) + status->offset_ino);

	if (check_family(newsk, AF_INET)) {
		struct tcp_ipv4_event_t evt = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_ACCEPT,
			.netns = net_ns_inum,
		};
		evt.pid = pid >> 32;
		bpf_probe_read(&evt.saddr, sizeof(u32), ((char *)newsk) + status->offset_saddr);
		bpf_probe_read(&evt.daddr, sizeof(u32), ((char *)newsk) + status->offset_daddr);

		evt.sport = lport;
		evt.dport = ntohs(dport);
		bpf_get_current_comm(&evt.comm, sizeof(evt.comm));

		// do not send event if IP address is 0.0.0.0 or port is 0
		if (evt.saddr != 0 && evt.daddr != 0 && evt.sport != 0 && evt.dport != 0) {
			bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt, sizeof(evt));
		}
	} else if (check_family(newsk, AF_INET6)) {
		struct tcp_ipv6_event_t evt = {
			.timestamp = bpf_ktime_get_ns(),
			.cpu = cpu,
			.type = TCP_EVENT_TYPE_ACCEPT,
			.netns = net_ns_inum,
		};
		evt.pid = pid >> 32;
		bpf_probe_read(&evt.daddr_h, sizeof(u64), ((char *)newsk) + status->offset_daddr_ipv6);
		bpf_probe_read(&evt.daddr_l, sizeof(u64), ((char *)newsk) + status->offset_daddr_ipv6 + sizeof(u64));
		bpf_probe_read(&evt.saddr_h, sizeof(u64), ((char *)newsk) + status->offset_daddr_ipv6 + 2 * sizeof(u64));
		bpf_probe_read(&evt.saddr_l, sizeof(u64), ((char *)newsk) + status->offset_daddr_ipv6 + 3 * sizeof(u64));

		evt.sport = lport;
		evt.dport = ntohs(dport);
		bpf_get_current_comm(&evt.comm, sizeof(evt.comm));
		if (is_ipv4_mapped_ipv6(evt.saddr_h, evt.saddr_l, evt.daddr_h, evt.daddr_l)) {
			struct tcp_ipv4_event_t evt4 = {
				.timestamp = bpf_ktime_get_ns(),
				.cpu = cpu,
				.type = TCP_EVENT_TYPE_ACCEPT,
				.pid = pid >> 32,
				.saddr = (u32)(evt.saddr_l >> 32),
				.daddr = (u32)(evt.daddr_l >> 32),
				.sport = evt.sport,
				.dport = evt.dport,
				.netns = net_ns_inum,
			};
			bpf_get_current_comm(&evt4.comm, sizeof(evt4.comm));
			if (evt4.saddr != 0 && evt4.daddr != 0 && evt4.sport != 0 && evt4.dport != 0) {
				bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt4, sizeof(evt4));
			}
			return 0;
		}
		// do not send event if IP address is :: or port is 0
		if ((evt.saddr_h || evt.saddr_l) && (evt.daddr_h || evt.daddr_l) && evt.sport != 0 && evt.dport != 0) {
			bpf_perf_event_output(ctx, &tcp_event_ipv6, cpu, &evt, sizeof(evt));
		}
	}
	return 0;
}

SEC("kprobe/fd_install")
int kprobe__fd_install(struct pt_regs *ctx)
{
	u64 pid = bpf_get_current_pid_tgid();
	u32 tgid = pid >> 32;
	unsigned long fd = (unsigned long) PT_REGS_PARM1(ctx);
	u32 *exists = NULL;

	exists = bpf_map_lookup_elem(&fdinstall_pids, &tgid);
	if (exists == NULL || !*exists)
		return 0;

	bpf_map_update_elem(&fdinstall_ret, &pid, &fd, BPF_ANY);

	return 0;
}

SEC("kretprobe/fd_install")
int kretprobe__fd_install(struct pt_regs *ctx)
{
	u64 pid = bpf_get_current_pid_tgid();
	unsigned long *fd;
	fd = bpf_map_lookup_elem(&fdinstall_ret, &pid);
	if (fd == NULL) {
		return 0;	// missed entry
	}
	bpf_map_delete_elem(&fdinstall_ret, &pid);

	u32 cpu = bpf_get_smp_processor_id();
	struct tcp_ipv4_event_t evt = {
		.timestamp = bpf_ktime_get_ns(),
		.cpu = cpu,
		.type = TCP_EVENT_TYPE_FD_INSTALL,
	};
	evt.pid = pid >> 32;
	evt.fd = *(__u32*)fd;
	bpf_get_current_comm(&evt.comm, sizeof(evt.comm));
	bpf_perf_event_output(ctx, &tcp_event_ipv4, cpu, &evt, sizeof(evt));

	return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by gobpf-elf-loader to set the current
// running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
