package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include "plugin.h"
//
// typedef struct {
//   const char *name;
//   plugin_shutdown_cb callback;
// } shutdown_callback_t;
// static shutdown_callback_t *shutdown_callbacks = NULL;
// static size_t shutdown_callbacks_num = 0;
//
// int plugin_register_shutdown(const char *name, plugin_shutdown_cb callback) {
//   shutdown_callback_t *ptr =
//       realloc(shutdown_callbacks,
//               (shutdown_callbacks_num + 1) * sizeof(*shutdown_callbacks));
//   if (ptr == NULL) {
//     return ENOMEM;
//   }
//   shutdown_callbacks = ptr;
//   shutdown_callbacks[shutdown_callbacks_num] = (shutdown_callback_t){
//       .name = name,
//       .callback = callback,
//   };
//   shutdown_callbacks_num++;
//
//   return 0;
// }
//
// int plugin_shutdown_all(void) {
//   int ret = 0;
//   for (size_t i = 0; i < shutdown_callbacks_num; i++) {
//     int err = shutdown_callbacks[i].callback();
//     if (err != 0) {
//       ret = err;
//     }
//   }
//   return ret;
// }
//
// void reset_shutdown(void) {
//   free(shutdown_callbacks);
//   shutdown_callbacks = NULL;
//   shutdown_callbacks_num = 0;
// }
import "C"

import (
	"fmt"
)

// ShutdownAll calls all registered shutdown callbacks.
func ShutdownAll() error {
	status, err := C.plugin_shutdown_all()
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("plugin_shutdown_all() = %d", status)
	}

	return nil
}
