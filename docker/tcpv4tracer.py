#!/usr/bin/python
#
# tcpv4tracer	Trace TCP IPv4 connections.
#		For Linux, uses BCC, eBPF. Embedded C.
#
# USAGE: tcpv4tracer [-h] [-p PID]
#
from __future__ import print_function
from bcc import BPF

import argparse
import ctypes

parser = argparse.ArgumentParser(
    description="Trace TCP IPv4 connections",
    formatter_class=argparse.RawDescriptionHelpFormatter)
parser.add_argument("-p", "--pid",
    help="trace this PID only")
args = parser.parse_args()

# define BPF program
bpf_text = """
#include <uapi/linux/ptrace.h>
#include <net/sock.h>
#include <net/inet_sock.h>
#include <net/net_namespace.h>
#include <bcc/proto.h>

struct tcp_event_t {
	char type[12];
	u32 pid;
	u32 saddr;
	u32 daddr;
	u16 sport;
	u16 dport;
};

BPF_PERF_OUTPUT(tcp_event);
BPF_HASH(connectsock, u64, struct sock *);
BPF_HASH(closesock, u64, struct sock *);

int kprobe__tcp_v4_connect(struct pt_regs *ctx, struct sock *sk)
{
	u64 pid = bpf_get_current_pid_tgid();

	##FILTER_PID##

	// stash the sock ptr for lookup on return
	connectsock.update(&pid, &sk);

	return 0;
};

int kretprobe__tcp_v4_connect(struct pt_regs *ctx)
{
	int ret = PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();

	struct sock **skpp;
	skpp = connectsock.lookup(&pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	if (ret != 0) {
		// failed to send SYNC packet, may not have populated
		// socket __sk_common.{skc_rcv_saddr, ...}
		connectsock.delete(&pid);
		return 0;
	}


	// pull in details
	struct sock *skp = *skpp;
	struct ns_common *ns;
	u32 saddr = 0, daddr = 0;
	u16 sport = 0, dport = 0;
	bpf_probe_read(&sport, sizeof(sport), &((struct inet_sock *)skp)->inet_sport);
	bpf_probe_read(&saddr, sizeof(saddr), &skp->__sk_common.skc_rcv_saddr);
	bpf_probe_read(&daddr, sizeof(daddr), &skp->__sk_common.skc_daddr);
	bpf_probe_read(&dport, sizeof(dport), &skp->__sk_common.skc_dport);

	// output
	struct tcp_event_t evt = {
		.type = "connect",
		.pid = pid >> 32,
		.saddr = saddr,
		.daddr = daddr,
		.sport = ntohs(sport),
		.dport = ntohs(dport),
	};

	u16 family = 0;
	bpf_probe_read(&family, sizeof(family), &skp->__sk_common.skc_family);

	tcp_event.perf_submit(ctx, &evt, sizeof(evt));

	connectsock.delete(&pid);

	return 0;
}

int kprobe__tcp_close(struct pt_regs *ctx, struct sock *sk)
{
	u64 pid = bpf_get_current_pid_tgid();

	##FILTER_PID##

	// stash the sock ptr for lookup on return
	closesock.update(&pid, &sk);

	return 0;
};

int kretprobe__tcp_close(struct pt_regs *ctx)
{
	u64 pid = bpf_get_current_pid_tgid();

	struct sock **skpp;
	skpp = closesock.lookup(&pid);
	if (skpp == 0) {
		return 0;	// missed entry
	}

	// pull in details
	struct sock *skp = *skpp;
	u32 saddr = 0, daddr = 0;
	u16 sport = 0, dport = 0;
	bpf_probe_read(&saddr, sizeof(saddr), &skp->__sk_common.skc_rcv_saddr);
	bpf_probe_read(&daddr, sizeof(daddr), &skp->__sk_common.skc_daddr);
	bpf_probe_read(&sport, sizeof(sport), &((struct inet_sock *)skp)->inet_sport);
	bpf_probe_read(&dport, sizeof(dport), &skp->__sk_common.skc_dport);

	// output
	struct tcp_event_t evt = {
		.type = "close",
		.pid = pid >> 32,
		.saddr = saddr,
		.daddr = daddr,
		.sport = ntohs(sport),
		.dport = ntohs(dport),
	};

	u16 family = 0;
	bpf_probe_read(&family, sizeof(family), &skp->__sk_common.skc_family);

	tcp_event.perf_submit(ctx, &evt, sizeof(evt));

	closesock.delete(&pid);

	return 0;
}

int kretprobe__inet_csk_accept(struct pt_regs *ctx)
{
	struct sock *newsk = (struct sock *)PT_REGS_RC(ctx);
	u64 pid = bpf_get_current_pid_tgid();

	##FILTER_PID##

	if (newsk == NULL)
		return 0;

	// check this is TCP
	u8 protocol = 0;
	// workaround for reading the sk_protocol bitfield:
	bpf_probe_read(&protocol, 1, (void *)((long)&newsk->sk_wmem_queued) - 3);
	if (protocol != IPPROTO_TCP)
		return 0;

	// pull in details
	u16 family = 0, lport = 0, dport = 0;
	bpf_probe_read(&family, sizeof(family), &newsk->__sk_common.skc_family);
	bpf_probe_read(&lport, sizeof(lport), &newsk->__sk_common.skc_num);
	bpf_probe_read(&dport, sizeof(dport), &newsk->__sk_common.skc_dport);

	if (family == AF_INET) {
		struct tcp_event_t evt = {.type = "accept", .pid = pid >> 32};
		bpf_probe_read(&evt.saddr, sizeof(u32),
			&newsk->__sk_common.skc_rcv_saddr);
		bpf_probe_read(&evt.daddr, sizeof(u32),
			&newsk->__sk_common.skc_daddr);
			evt.sport = lport;
		evt.dport = ntohs(dport);
		tcp_event.perf_submit(ctx, &evt, sizeof(evt));
	}
	// else drop

	return 0;
}
"""

TASK_COMM_LEN = 16   # linux/sched.h
class TCPEvt(ctypes.Structure):
	_fields_ = [
		("type", ctypes.c_char * 12),
		("pid", ctypes.c_uint),
		("saddr", ctypes.c_uint),
		("daddr", ctypes.c_uint),
		("sport", ctypes.c_ushort),
		("dport", ctypes.c_ushort),
	]

def print_event(cpu, data, size):
	event = ctypes.cast(data, ctypes.POINTER(TCPEvt)).contents
	print("%-12s %-6s %-16s %-16s %-16s %-6s %-6s" % (event.type.decode('utf-8'), event.pid, " ",
	    inet_ntoa(event.saddr),
	    inet_ntoa(event.daddr),
	    event.sport,
	    event.dport,
	    ))

if args.pid:
    bpf_text = bpf_text.replace('##FILTER_PID##',
        'if (pid != %s) { return 0; }' % args.pid)
else:
    bpf_text = bpf_text.replace('##FILTER_PID##', '')

# initialize BPF
b = BPF(text=bpf_text)

# header
print("%-12s %-6s %-16s %-16s %-16s %-6s %-6s" % ("TYPE", "PID", "COMM", "SADDR", "DADDR",
    "SPORT", "DPORT"))

def inet_ntoa(addr):
	dq = ''
	for i in range(0, 4):
		dq = dq + str(addr & 0xff)
		if (i != 3):
			dq = dq + '.'
		addr = addr >> 8
	return dq

b["tcp_event"].open_perf_buffer(print_event)
while True:
	b.kprobe_poll()
