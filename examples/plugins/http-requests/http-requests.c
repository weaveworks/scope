#include <linux/skbuff.h>
#include <net/sock.h>

/* Table from (Task group id|Task id) to (Number of received http requests).
   We need to gather requests per task and not only per task group (i.e. userspace pid)
   so that entries can be cleared up independently when a task exists.
   This implies that userspace needs to do the per-process aggregation.
 */
BPF_HASH(received_http_requests, u64, u64);


/* skb_copy_datagram_iter() (Kernels >= 3.19) is in charge of copying socket
   buffers from kernel to userspace.

   skb_copy_datagram_iter() has an associated tracepoint
   (trace_skb_copy_datagram_iovec), which would be more stable than a kprobe but
   it lacks the offset argument.
 */
int kprobe__skb_copy_datagram_iter(struct pt_regs *ctx, const struct sk_buff *skb, int offset, void *unused_iovec, int len)
{

  /* Inspect the beginning of socket buffers copied to user-space to determine
     if they correspond to http requests.

     Caveats:

     Requests may not appear at the beginning of a packet due to:
     * Persistent connections.
     * Packet fragmentation.

     We could inspect the full packet but:
     * It's very inefficient.
     * Examining the non-linear (paginated) area of a socket buffer would be
       really tricky from ebpf.
  */

  /* Verify it's a TCP socket
     TODO: is it worth caching it in a socket table?
   */
  struct sock *sk = skb->sk;
  unsigned short skc_family = sk->__sk_common.skc_family;
  switch (skc_family) {
    case PF_INET:
    case PF_INET6:
    case PF_UNIX:
      break;
    default:
      return 0;
  }
  /* The socket type and protocol are not directly addressable since they are
     bitfields.  We access them by assuming sk_write_queue is immediately before
     them (admittedly pretty hacky).
  */
  unsigned int flags = 0;
  size_t flags_offset = offsetof(typeof(struct sock), sk_write_queue) + sizeof(sk->sk_write_queue);
  bpf_probe_read(&flags, sizeof(flags), ((u8*)sk) + flags_offset);
  u16 sk_type = flags >> 16;
  if (sk_type != SOCK_STREAM) {
    return 0;
  }
  u8 sk_protocol = flags >> 8 & 0xFF;
  /* The protocol is unset (IPPROTO_IP) in Unix sockets */
  if ( (sk_protocol != IPPROTO_TCP) && ((skc_family == PF_UNIX) && (sk_protocol != IPPROTO_IP)) ) {
    return 0;
  }

  /* Inline implementation of skb_headlen() */
  unsigned int head_len = skb->len - skb->data_len;
  /* http://stackoverflow.com/questions/25047905/http-request-minimum-size-in-bytes
     minimum length of http request is always greater than 7 bytes
  */
  unsigned int available_data = head_len - offset;
  if (available_data < 7) {
    return 0;
  }

  /* Check if buffer begins with a method name followed by a space.

     To avoid false positives it would be good to do a deeper inspection
     (i.e. fully ensure a 'Method SP Request-URI SP HTTP-Version CRLF'
     structure) but loops are not allowed in ebpf, making variable-size-data
     parsers infeasible.
  */
  u8 data[8] = {};
  if (available_data >= 8) {
    /* We have confirmed having access to 7 bytes, but need 8 bytes to check the
       space after OPTIONS. bpf_probe_read() requires its second argument to be
       an immediate, so we obtain the data in this unsexy way.
    */
    bpf_probe_read(&data, 8, skb->data + offset);
  } else {
    bpf_probe_read(&data, 7, skb->data + offset);
  }

  switch (data[0]) {
    /* DELETE */
    case 'D':
      if ((data[1] != 'E') || (data[2] != 'L') || (data[3] != 'E') || (data[4] != 'T') || (data[5] != 'E') || (data[6] != ' ')) {
        return 0;
      }
      break;

    /* GET */
    case 'G':
      if ((data[1] != 'E') || (data[2] != 'T') || (data[3] != ' ')) {
        return 0;
      }
      break;

    /* HEAD */
    case 'H':
      if ((data[1] != 'E') || (data[2] != 'A') || (data[3] != 'D') || (data[4] != ' ')) {
        return 0;
      }
      break;

    /* OPTIONS */
    case 'O':
      if (available_data < 8 || (data[1] != 'P') || (data[2] != 'T') || (data[3] != 'I') || (data[4] != 'O') || (data[5] != 'N') || (data[6] != 'S') || (data[7] != ' ')) {
        return 0;
      }
      break;

    /* PATCH/POST/PUT */
    case 'P':
      switch (data[1]) {
          case 'A':
            if ((data[2] != 'T') || (data[3] != 'C') || (data[4] != 'H') || (data[5] != ' ')) {
              return 0;
            }
            break;
          case 'O':
            if ((data[2] != 'S') || (data[3] != 'T') || (data[4] != ' ')) {
              return 0;
            }
            break;
          case 'U':
            if ((data[2] != 'T') || (data[3] != ' ')) {
              return 0;
            }
            break;
      }
      break;

    default:
      return 0;
  }

  /* Finally, bump the request counter for current task */
  u64 pid_tgid = bpf_get_current_pid_tgid();
  received_http_requests.increment(pid_tgid);

  return 0;
}


/* Clear out request count entries of tasks on exit */
int kprobe__do_exit(struct pt_regs *ctx) {
  u64 pid_tgid = bpf_get_current_pid_tgid();
  received_http_requests.delete(&pid_tgid);
  return 0;
}
