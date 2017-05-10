#ifndef __TCPTRACER_BPF_H
#define __TCPTRACER_BPF_H

#include <linux/types.h>

#define TCP_EVENT_TYPE_CONNECT          1
#define TCP_EVENT_TYPE_ACCEPT           2
#define TCP_EVENT_TYPE_CLOSE            3
#define TCP_EVENT_TYPE_FD_INSTALL       4

#define GUESS_SADDR      0
#define GUESS_DADDR      1
#define GUESS_FAMILY     2
#define GUESS_SPORT      3
#define GUESS_DPORT      4
#define GUESS_NETNS      5
#define GUESS_DADDR_IPV6 6

#ifndef TASK_COMM_LEN
#define TASK_COMM_LEN 16
#endif

struct tcp_ipv4_event_t {
	__u64 timestamp;
	__u64 cpu;
	__u32 type;
	__u32 pid;
	char comm[TASK_COMM_LEN];
	__u32 saddr;
	__u32 daddr;
	__u16 sport;
	__u16 dport;
	__u32 netns;
	__u32 fd;
	__u32 dummy;
};

struct tcp_ipv6_event_t {
	__u64 timestamp;
	__u64 cpu;
	__u32 type;
	__u32 pid;
	char comm[TASK_COMM_LEN];
	/* Using the type unsigned __int128 generates an error in the ebpf verifier */
	__u64 saddr_h;
	__u64 saddr_l;
	__u64 daddr_h;
	__u64 daddr_l;
	__u16 sport;
	__u16 dport;
	__u32 netns;
	__u32 fd;
	__u32 dummy;
};

// tcp_set_state doesn't run in the context of the process that initiated the
// connection so we need to store a map TUPLE -> PID to send the right PID on
// the event
struct ipv4_tuple_t {
	__u32 saddr;
	__u32 daddr;
	__u16 sport;
	__u16 dport;
	__u32 netns;
};

struct ipv6_tuple_t {
	/* Using the type unsigned __int128 generates an error in the ebpf verifier */
	__u64 saddr_h;
	__u64 saddr_l;
	__u64 daddr_h;
	__u64 daddr_l;
	__u16 sport;
	__u16 dport;
	__u32 netns;
};

struct pid_comm_t {
	__u64 pid;
	char comm[TASK_COMM_LEN];
};

#define TCPTRACER_STATE_UNINITIALIZED 0
#define TCPTRACER_STATE_CHECKING      1
#define TCPTRACER_STATE_CHECKED       2
#define TCPTRACER_STATE_READY         3
struct tcptracer_status_t {
	__u64 state;

	/* checking */
	__u64 pid_tgid;
	__u64 what;
	__u64 offset_saddr;
	__u64 offset_daddr;
	__u64 offset_sport;
	__u64 offset_dport;
	__u64 offset_netns;
	__u64 offset_ino;
	__u64 offset_family;
	__u64 offset_daddr_ipv6;

	__u64 err;

	__u32 daddr_ipv6[4];
	__u32 netns;
	__u32 saddr;
	__u32 daddr;
	__u16 sport;
	__u16 dport;
	__u16 family;
	__u16 padding;
};

#endif
