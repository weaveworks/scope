#include <uapi/linux/ptrace.h>
#include <linux/skbuff.h>

struct received_http_requests_key_t {
  u32 pid;
};
BPF_HASH(received_http_requests, struct received_http_requests_key_t, u64);


/*
  skb_copy_datagram_iter() (Kernels >= 3.19) is in charge of copying socket
  buffers from kernel to userspace.

  skb_copy_datagram_iter() has an associated tracepoint
  (trace_skb_copy_datagram_iovec), which would be more stable than a kprobe but
  it lacks the offset argument.
 */
int kprobe__skb_copy_datagram_iter(struct pt_regs *ctx, const struct sk_buff *skb, int offset, void *unused_iovec, int len)
{

  /* Inspect the beginning of socket buffers copied to user-space to determine if they
     correspond to http requests.

     Caveats:

     Requests may not appear at the beginning of a packet due to:
     * Persistent connections.
     * Packet fragmentation.

     We could inspect the full packet but:
     * It's very inefficient.
     * Examining the non-linear (paginated) area of a socket buffer would be
     really tricky from ebpf.
   */

  /* TODO: exit early if it's not TCP */

  /* Inline implementation of skb_headlen() */
  unsigned int head_len = skb->len - skb->data_len;

  /* http://stackoverflow.com/questions/25047905/http-request-minimum-size-in-bytes
     minimum length of http request is always geater than 7 bytes
  */
  if (head_len - offset < 7) {
    return 0;
  }

  u8 data[4] = {};
  bpf_probe_read(&data, sizeof(data), skb->data + offset);

  /* TODO: support other methods and optimize lookups */
  if ((data[0] == 'G') && (data[1] == 'E') && (data[2] == 'T') && (data[3] == ' ')) {
    /* Record request */
    struct received_http_requests_key_t key = {};
    key.pid = bpf_get_current_pid_tgid() >> 32;
    received_http_requests.increment(key);
  }


  return 0;
}
