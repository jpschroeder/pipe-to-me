#include <stdio.h>
#include <stdbool.h> 
#include <unistd.h> 
#include <fcntl.h>
#include <errno.h>
#include <string.h>
#include <curl/curl.h>

struct transfer_state {
  bool readbusy;
  CURL *curl;
};

static size_t tool_read_cb(void *buffer, size_t sz, size_t nmemb, void *data)
{
  ssize_t rc;
  struct transfer_state *state = data;

  rc = read(STDIN_FILENO, buffer, sz*nmemb);
  if(rc < 0) {
    if(errno == EAGAIN) {
      errno = 0;
      state->readbusy = true;
      return CURL_READFUNC_PAUSE;
    }
    rc = 0;
  }
  state->readbusy = false;
  return (size_t)rc;
}

static int xferinfo(void *data,
                    curl_off_t dltotal, curl_off_t dlnow,
                    curl_off_t ultotal, curl_off_t ulnow)
{
  struct transfer_state *state = data;
  if(state->readbusy) {
    state->readbusy = false;
    curl_easy_pause(state->curl, CURLPAUSE_CONT);
  }
  return 0;
}


int main(int argc, char* argv[])
{
  CURLcode res;
  struct transfer_state state = { false, 0 };

  if (argc < 2 || !strcmp(argv[1], "-h") || !strcmp(argv[1], "--help")) {
    fprintf(stderr, "usage: client https://pipeto.me/<code>\n");
    return -1;
  }

  if (fcntl(STDIN_FILENO, F_SETFL, O_NONBLOCK) < 0) {
    fprintf(stderr, "fcntl() failed\n");
    return -1;
  }

  state.curl = curl_easy_init();

  if(!state.curl) {
    fprintf(stderr, "curl_easy_init() failed\n");
    return -1;
  }

  curl_easy_setopt(state.curl, CURLOPT_URL, argv[1]);
  curl_easy_setopt(state.curl, CURLOPT_UPLOAD, 1L);
  curl_easy_setopt(state.curl, CURLOPT_READFUNCTION, tool_read_cb);
  curl_easy_setopt(state.curl, CURLOPT_READDATA, &state);
  curl_easy_setopt(state.curl, CURLOPT_XFERINFOFUNCTION, xferinfo);
  curl_easy_setopt(state.curl, CURLOPT_XFERINFODATA, &state);
  curl_easy_setopt(state.curl, CURLOPT_NOPROGRESS, 0L);

  printf("connected to: %s\n", argv[1]);
  res = curl_easy_perform(state.curl);
  if(res != CURLE_OK)
    fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));

  curl_easy_cleanup(state.curl);

  return 0;
}