// +build go1.5,cgo

package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include "plugin.h"
// #include <stdlib.h>
// #include <dlfcn.h>
//
// int (*register_read_ptr) (char const *group, char const *name,
//     plugin_read_cb callback,
//     cdtime_t interval,
//     user_data_t *ud) = NULL;
// int register_read_wrapper (char const *group, char const *name,
//     plugin_read_cb callback,
//     cdtime_t interval,
//     user_data_t *ud) {
//   if (register_read_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_read_ptr = dlsym(hnd, "plugin_register_complex_read");
//     dlclose(hnd);
//   }
//   return (*register_read_ptr) (group, name, callback, interval, ud);
// }
//
// int (*dispatch_values_ptr) (value_list_t const *vl);
// int dispatch_values_wrapper (value_list_t const *vl) {
//   if (dispatch_values_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     dispatch_values_ptr = dlsym(hnd, "plugin_dispatch_values");
//     dlclose(hnd);
//   }
//   return (*dispatch_values_ptr) (vl);
// }
//
// void value_list_add (value_list_t *vl, value_t v) {
//   value_t *tmp;
//   tmp = realloc (vl->values, (vl->values_len + 1));
//   if (tmp == NULL) {
//     errno = ENOMEM;
//     return;
//   }
//   vl->values = tmp;
//   vl->values[vl->values_len] = v;
//   vl->values_len++;
// }
//
// data_source_t *ds_dsrc(data_set_t const *ds, size_t i) { return ds->ds + i; }
//
// void value_list_add_counter (value_list_t *vl, counter_t c) {
//   value_list_add (vl, (value_t){
//     .counter = c,
//   });
// }
//
// void value_list_add_gauge (value_list_t *vl, gauge_t g) {
//   value_list_add (vl, (value_t){
//     .gauge = g,
//   });
// }
//
// void value_list_add_derive (value_list_t *vl, derive_t d) {
//   value_list_add (vl, (value_t){
//     .derive = d,
//   });
// }
//
// counter_t value_list_get_counter (value_list_t *vl, size_t i) {
//   return vl->values[i].counter;
// }
//
// gauge_t value_list_get_gauge (value_list_t *vl, size_t i) {
//   return vl->values[i].gauge;
// }
//
// derive_t value_list_get_derive (value_list_t *vl, size_t i) {
//   return vl->values[i].derive;
// }
//
// int (*register_write_ptr) (char const *, plugin_write_cb, user_data_t *);
// int register_write_wrapper (char const *name, plugin_write_cb callback, user_data_t *user_data) {
//   if (register_write_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_write_ptr = dlsym(hnd, "plugin_register_write");
//     dlclose(hnd);
//   }
//   return (*register_write_ptr) (name, callback, user_data);
// }
//
// int (*register_shutdown_ptr) (char *, plugin_shutdown_cb);
// int register_shutdown_wrapper (char *name, plugin_shutdown_cb callback) {
//   if (register_shutdown_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_shutdown_ptr = dlsym(hnd, "plugin_register_shutdown");
//     dlclose(hnd);
//   }
//   return (*register_shutdown_ptr) (name, callback);
//
// }
//
// meta_data_t *(*meta_data_create_ptr) (void) = NULL;
// meta_data_t *meta_data_create_wrapper(void) {
//   if (meta_data_create_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     meta_data_create_ptr = dlsym(hnd, "meta_data_create");
//     dlclose(hnd);
//   }
//   return (*meta_data_create_ptr)();
// }
//
// meta_data_t *(*meta_data_destroy_ptr) (meta_data_t *meta) = NULL;
// meta_data_t *meta_data_destroy_wrapper(meta_data_t *meta) {
//   if (meta_data_destroy_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     meta_data_destroy_ptr = dlsym(hnd, "meta_data_destroy");
//     dlclose(hnd);
//   }
//   return (*meta_data_destroy_ptr)(meta);
// }
//
// int (*meta_data_add_string_ptr) (meta_data_t *md,
//   const char *key, const char *value) = NULL;
// int meta_data_add_string_wrapper(meta_data_t *md,
//   const char *key, const char *value) {
//     if (meta_data_add_string_ptr == NULL) {
//       void *hnd = dlopen(NULL, RTLD_LAZY);
//       meta_data_add_string_ptr = dlsym(hnd, "meta_data_add_string");
//       dlclose(hnd);
//     }
//     return (*meta_data_add_string_ptr)(md, key, value);
// }
//
// int (*meta_data_add_signed_int_ptr) (meta_data_t *md,
//   const char *key, int64_t value) = NULL;
// int meta_data_add_signed_int_wrapper(meta_data_t *md,
//   const char *key, int64_t value) {
//     if (meta_data_add_signed_int_ptr == NULL) {
//       void *hnd = dlopen(NULL, RTLD_LAZY);
//       meta_data_add_signed_int_ptr = dlsym(hnd, "meta_data_add_signed_int");
//       dlclose(hnd);
//     }
//     return (*meta_data_add_signed_int_ptr)(md, key, value);
// }
//
// int (*meta_data_add_unsigned_int_ptr) (meta_data_t *md,
//   const char *key, uint64_t value) = NULL;
// int meta_data_add_unsigned_int_wrapper(meta_data_t *md,
//   const char *key, uint64_t value) {
//     if (meta_data_add_unsigned_int_ptr == NULL) {
//       void *hnd = dlopen(NULL, RTLD_LAZY);
//       meta_data_add_unsigned_int_ptr = dlsym(hnd, "meta_data_add_unsigned_int");
//       dlclose(hnd);
//     }
//     return (*meta_data_add_unsigned_int_ptr)(md, key, value);
// }
//
// int (*meta_data_add_double_ptr) (meta_data_t *md,
//   const char *key, double value) = NULL;
// int meta_data_add_double_wrapper(meta_data_t *md,
//   const char *key, double value) {
//     if (meta_data_add_double_ptr == NULL) {
//       void *hnd = dlopen(NULL, RTLD_LAZY);
//       meta_data_add_double_ptr = dlsym(hnd, "meta_data_add_double");
//       dlclose(hnd);
//     }
//     return (*meta_data_add_double_ptr)(md, key, value);
// }
//
// int (*meta_data_add_boolean_ptr) (meta_data_t *md,
//   const char *key, _Bool value) = NULL;
// int meta_data_add_boolean_wrapper(meta_data_t *md,
//   const char *key, _Bool value) {
//     if (meta_data_add_boolean_ptr == NULL) {
//       void *hnd = dlopen(NULL, RTLD_LAZY);
//       meta_data_add_boolean_ptr = dlsym(hnd, "meta_data_add_boolean");
//       dlclose(hnd);
//     }
//     return (*meta_data_add_boolean_ptr)(md, key, value);
// }
import "C"
