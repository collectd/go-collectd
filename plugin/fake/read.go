package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include "plugin.h"
//
// typedef struct {
//   const char *group;
//   const char *name;
//   plugin_read_cb callback;
//   cdtime_t interval;
//   user_data_t user_data;
// } read_callback_t;
// static read_callback_t *read_callbacks = NULL;
// static size_t read_callbacks_num = 0;
//
// int plugin_register_complex_read(const char *group, const char *name,
//                                  plugin_read_cb callback, cdtime_t interval,
//                                  user_data_t const *user_data) {
//   read_callback_t *ptr = realloc(
//       read_callbacks, (read_callbacks_num + 1) * sizeof(*read_callbacks));
//   if (ptr == NULL) {
//     return ENOMEM;
//   }
//   read_callbacks = ptr;
//   read_callbacks[read_callbacks_num] = (read_callback_t){
//       .group = group,
//       .name = name,
//       .callback = callback,
//       .interval = interval,
//       .user_data = *user_data,
//   };
//   read_callbacks_num++;
//
//   return 0;
// }
//
// void plugin_set_interval(cdtime_t);
// static int read_all(void) {
//   cdtime_t save_interval = plugin_get_interval();
//   int ret = 0;
//
//   for (size_t i = 0; i < read_callbacks_num; i++) {
//     read_callback_t *cb = read_callbacks + i;
//     plugin_set_interval(cb->interval);
//     int err = cb->callback(&cb->user_data);
//     if (err != 0) {
//       ret = err;
//     }
//   }
//
//   plugin_set_interval(save_interval);
//   return ret;
// }
//
// void reset_read(void) {
//   for (size_t i = 0; i < read_callbacks_num; i++) {
//     user_data_t *ud = &read_callbacks[i].user_data;
//     if (ud->free_func == NULL) {
//       continue;
//     }
//     ud->free_func(ud->data);
//     ud->data = NULL;
//   }
//   free(read_callbacks);
//   read_callbacks = NULL;
//   read_callbacks_num = 0;
// }
import "C"

import (
	"fmt"
)

func ReadAll() error {
	status, err := C.read_all()
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("read_all() = %d", status)
	}

	return nil
}
