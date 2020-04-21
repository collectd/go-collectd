package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include "plugin.h"
//
// typedef struct {
//   const char *name;
//   plugin_log_cb callback;
//   user_data_t user_data;
// } log_callback_t;
// static log_callback_t *log_callbacks = NULL;
// static size_t log_callbacks_num = 0;
//
// int plugin_register_log(const char *name,
//                         plugin_log_cb callback,
//                         user_data_t const *user_data) {
//   log_callback_t *ptr = realloc(log_callbacks, (log_callbacks_num+1) * sizeof(*log_callbacks));
//   if (ptr == NULL) {
//     return ENOMEM;
//   }
//   log_callbacks = ptr;
//   log_callbacks[log_callbacks_num] = (log_callback_t){
//     .name = name,
//     .callback = callback,
//     .user_data = *user_data,
//   };
//   log_callbacks_num++;
//
//   return 0;
// }
//
// void plugin_log(int level, const char *format, ...) {
//   char msg[1024];
//   va_list ap;
//   va_start(ap, format);
//   vsnprintf(msg, sizeof(msg), format, ap);
//   msg[sizeof(msg)-1] = 0;
//   va_end(ap);
//
//   for (size_t i = 0; i < log_callbacks_num; i++) {
//     log_callbacks[i].callback(level, msg, &log_callbacks[i].user_data);
//   }
// }
//
// void reset_log(void) {
//   free(log_callbacks);
//   log_callbacks = NULL;
//   log_callbacks_num = 0;
// }
import "C"
