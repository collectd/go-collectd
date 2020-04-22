package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <stdio.h>
// #include "plugin.h"
//
// typedef struct {
//   const char *name;
//   plugin_write_cb callback;
//   user_data_t user_data;
// } write_callback_t;
// static write_callback_t *write_callbacks = NULL;
// static size_t write_callbacks_num = 0;
//
// int plugin_register_write(const char *name, plugin_write_cb callback,
//                           user_data_t const *user_data) {
//   write_callback_t *ptr = realloc(
//       write_callbacks, (write_callbacks_num + 1) * sizeof(*write_callbacks));
//   if (ptr == NULL) {
//     return ENOMEM;
//   }
//   write_callbacks = ptr;
//   write_callbacks[write_callbacks_num] = (write_callback_t){
//       .name = name,
//       .callback = callback,
//       .user_data = *user_data,
//   };
//   write_callbacks_num++;
//
//   return 0;
// }
//
// int plugin_dispatch_values(value_list_t const *vl) {
//   data_set_t *ds = &(data_set_t){
//       .ds_num = 1,
//       .ds =
//           &(data_source_t){
//               .name = "value",
//               .min = 0,
//               .max = NAN,
//           },
//   };
//
//   if (strcmp("derive", vl->type) == 0) {
//     strncpy(ds->type, vl->type, sizeof(ds->type));
//     ds->ds[0].type = DS_TYPE_DERIVE;
//   } else if (strcmp("gauge", vl->type) == 0) {
//     strncpy(ds->type, vl->type, sizeof(ds->type));
//     ds->ds[0].type = DS_TYPE_GAUGE;
//   } else if (strcmp("counter", vl->type) == 0) {
//     strncpy(ds->type, vl->type, sizeof(ds->type));
//     ds->ds[0].type = DS_TYPE_COUNTER;
//   } else {
//     errno = EINVAL;
//     return errno;
//   }
//
//   int ret = 0;
//   for (size_t i = 0; i < write_callbacks_num; i++) {
//     int err =
//         write_callbacks[i].callback(ds, vl, &write_callbacks[i].user_data);
//     if (err != 0) {
//       ret = err;
//     }
//   }
//
//   return ret;
// }
//
// void reset_write(void) {
//   free(write_callbacks);
//   write_callbacks = NULL;
//   write_callbacks_num = 0;
// }
import "C"
