package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <string.h>
// #include "plugin.h"
//
// typedef struct {
//   char *group;
//   char *name;
//   plugin_read_cb callback;
//   cdtime_t interval;
//   user_data_t user_data;
// } read_callback_t;
// read_callback_t *read_callbacks = NULL;
// size_t read_callbacks_num = 0;
//
// int plugin_register_complex_read(const char *group, const char *name,
//                                  plugin_read_cb callback, cdtime_t interval,
//                                  user_data_t const *user_data) {
//   if (interval == 0) {
//     interval = plugin_get_interval();
//   }
//
//   read_callback_t *ptr = realloc(
//       read_callbacks, (read_callbacks_num + 1) * sizeof(*read_callbacks));
//   if (ptr == NULL) {
//     return ENOMEM;
//   }
//   read_callbacks = ptr;
//   read_callbacks[read_callbacks_num] = (read_callback_t){
//       .group = (group != NULL) ? strdup(group) : NULL,
//       .name = strdup(name),
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
//     free(read_callbacks[i].name);
//     free(read_callbacks[i].group);
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
	"unsafe"

	"collectd.org/cdtime"
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

// ReadCallback represents a data associated with a registered read callback.
type ReadCallback struct {
	Group, Name string
	Interval    cdtime.Time
}

// ReadCallbacks returns the data associated with all registered read
// callbacks.
func ReadCallbacks() []ReadCallback {
	var ret []ReadCallback

	for i := C.size_t(0); i < C.read_callbacks_num; i++ {
		// Go pointer arithmetic that does the equivalent of C's `read_callbacks[i]`.
		cb := (*C.read_callback_t)(unsafe.Pointer(uintptr(unsafe.Pointer(C.read_callbacks)) + uintptr(C.sizeof_read_callback_t*i)))
		ret = append(ret, ReadCallback{
			Group:    C.GoString(cb.group),
			Name:     C.GoString(cb.name),
			Interval: cdtime.Time(cb.interval),
		})
	}

	return ret
}
