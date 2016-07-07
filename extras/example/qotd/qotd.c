#include <pthread.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <sys/syscall.h>

/* this function is run by the  thread */
void *thread_func(void *sock) {
  struct sockaddr_in dest;
  char in_buffer[1024];
  char out_buffer[1024];
  int sockfd = (int)sock;
  int clientfd;

  printf("I'm thread %d\n", syscall(SYS_gettid));

  clientfd = socket(AF_INET, SOCK_STREAM, 0);
  if (clientfd < 0) {
    perror("ERROR opening socket");
    return;
  }

  dest.sin_family = AF_INET;
  dest.sin_port = htons(17);
  if (inet_aton("104.230.14.102", &dest.sin_addr.s_addr) == 0) {
    perror("cygnus");
    return;
  }

  if (connect(clientfd, (struct sockaddr*)&dest, sizeof(dest)) != 0 ) {
    perror("Connect ");
    return;
  }

  int readbytes = recv(clientfd, in_buffer, sizeof(in_buffer), 0);
  close(clientfd);

  int writtenbytes = snprintf(out_buffer, sizeof(out_buffer), "{\"qotd\": %s}\n", in_buffer);
  if (write(sockfd, out_buffer, writtenbytes) < 0) {
    perror("ERROR writing to socket");
    return;
  }

  if (close(sockfd)) {
    perror("ERROR closing socket");
    return;
  }
}

int main(int argc, char **argv) {
  int sockfd, newsockfd, portno, clilen;
  struct sockaddr_in serv_addr, cli_addr;
  int n;

  sockfd = socket(AF_INET, SOCK_STREAM, 0);
  if (sockfd < 0) {
    perror("ERROR opening socket");
    return 1;
  }

  int i = 1;
  if (setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &i, sizeof(i)) == -1) {
      perror("setsockopt");
  }

  // bzero((char *) &serv_addr, sizeof(serv_addr));
  portno = 4446;
  serv_addr.sin_family = AF_INET;
  serv_addr.sin_addr.s_addr = INADDR_ANY;
  serv_addr.sin_port = htons(portno);

  if (bind(sockfd, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) < 0) {
    perror("ERROR on binding");
    return 1;
  }

  listen(sockfd, 5);

  while (1) {
    clilen = sizeof(cli_addr);
    newsockfd = accept(sockfd, (struct sockaddr *)&cli_addr, &clilen);
    if (newsockfd < 0) {
      perror("ERROR on accept");
      return 1;
    }

    pthread_t thread0;
    if(pthread_create(&thread0, NULL, thread_func, (void *)newsockfd)) {
        perror("ERROR on pthread_create");
        return 1;
    }
  }
  return 0;
}

