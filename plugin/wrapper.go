// +build go1.5,cgo

package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include "plugin.h"
// #include <dlfcn.h>
// #include <stdbool.h>
// #include <stdlib.h>
//
// #define LOAD(f)                                                                \
//   if (f##_ptr == NULL) {                                                       \
//     void *hnd = dlopen(NULL, RTLD_LAZY);                                       \
//     f##_ptr = dlsym(hnd, #f);                                                  \
//     dlclose(hnd);                                                              \
//   }
//
// static int (*meta_data_add_boolean_ptr)(meta_data_t *, char const *, bool);
// static int (*meta_data_add_double_ptr)(meta_data_t *, char const *, double);
// static int (*meta_data_add_signed_int_ptr)(meta_data_t *, char const *,
//                                            int64_t);
// static int (*meta_data_add_string_ptr)(meta_data_t *, char const *,
//                                        char const *);
// static int (*meta_data_add_unsigned_int_ptr)(meta_data_t *, char const *,
//                                              uint64_t);
// static meta_data_t *(*meta_data_create_ptr)(void);
// static void (*meta_data_destroy_ptr)(meta_data_t *);
// static int (*meta_data_get_boolean_ptr)(meta_data_t *, char const *, bool *);
// static int (*meta_data_get_double_ptr)(meta_data_t *, char const *, double *);
// static int (*meta_data_get_signed_int_ptr)(meta_data_t *, char const *,
//                                            int64_t *);
// static int (*meta_data_get_string_ptr)(meta_data_t *, char const *, char **);
// static int (*meta_data_get_unsigned_int_ptr)(meta_data_t *, char const *,
//                                              uint64_t *);
// static int (*meta_data_toc_ptr)(meta_data_t *, char ***);
// static int (*meta_data_type_ptr)(meta_data_t *, char const *);
// static int (*plugin_dispatch_values_ptr)(value_list_t const *);
// static cdtime_t (*plugin_get_interval_ptr)(void);
// static int (*plugin_register_complex_read_ptr)(meta_data_t *, char const *,
//                                                plugin_read_cb, cdtime_t,
//                                                user_data_t *);
// static int (*plugin_register_log_ptr)(char const *, plugin_log_cb,
//                                       user_data_t *);
// static int (*plugin_register_shutdown_ptr)(char const *, plugin_shutdown_cb);
// static int (*plugin_register_write_ptr)(char const *, plugin_write_cb,
//                                         user_data_t *);
//
// int meta_data_add_boolean_wrapper(meta_data_t *md, char const *key,
//                                   bool value) {
//   LOAD(meta_data_add_boolean);
//   return (*meta_data_add_boolean_ptr)(md, key, value);
// }
//
// int meta_data_add_double_wrapper(meta_data_t *md, char const *key,
//                                  double value) {
//   LOAD(meta_data_add_double);
//   return (*meta_data_add_double_ptr)(md, key, value);
// }
//
// int meta_data_add_signed_int_wrapper(meta_data_t *md, char const *key,
//                                      int64_t value) {
//   LOAD(meta_data_add_signed_int);
//   return (*meta_data_add_signed_int_ptr)(md, key, value);
// }
//
// int meta_data_add_string_wrapper(meta_data_t *md, char const *key,
//                                  char const *value) {
//   LOAD(meta_data_add_string);
//   return (*meta_data_add_string_ptr)(md, key, value);
// }
//
// int meta_data_add_unsigned_int_wrapper(meta_data_t *md, char const *key,
//                                        uint64_t value) {
//   LOAD(meta_data_add_unsigned_int);
//   return (*meta_data_add_unsigned_int_ptr)(md, key, value);
// }
//
// meta_data_t *meta_data_create_wrapper(void) {
//   LOAD(meta_data_create);
//   return (*meta_data_create_ptr)();
// }
//
// void meta_data_destroy_wrapper(meta_data_t *md) {
//   LOAD(meta_data_destroy);
//   (*meta_data_destroy_ptr)(md);
// }
//
// int meta_data_get_boolean_wrapper(meta_data_t *md, char const *key,
//                                   bool *value) {
//   LOAD(meta_data_get_boolean);
//   return (*meta_data_get_boolean_ptr)(md, key, value);
// }
//
// int meta_data_get_double_wrapper(meta_data_t *md, char const *key,
//                                  double *value) {
//   LOAD(meta_data_get_double);
//   return (*meta_data_get_double_ptr)(md, key, value);
// }
//
// int meta_data_get_signed_int_wrapper(meta_data_t *md, char const *key,
//                                      int64_t *value) {
//   LOAD(meta_data_get_signed_int);
//   return (*meta_data_get_signed_int_ptr)(md, key, value);
// }
//
// int meta_data_get_string_wrapper(meta_data_t *md, char const *key,
//                                  char **value) {
//   LOAD(meta_data_get_string);
//   return (*meta_data_get_string_ptr)(md, key, value);
// }
//
// int meta_data_get_unsigned_int_wrapper(meta_data_t *md, char const *key,
//                                        uint64_t *value) {
//   LOAD(meta_data_get_unsigned_int);
//   return (*meta_data_get_unsigned_int_ptr)(md, key, value);
// }
//
// int meta_data_toc_wrapper(meta_data_t *md, char ***toc) {
//   LOAD(meta_data_toc);
//   return (*meta_data_toc_ptr)(md, toc);
// }
//
// int meta_data_type_wrapper(meta_data_t *md, char const *key) {
//   LOAD(meta_data_type);
//   return (*meta_data_type_ptr)(md, key);
// }
//
// int plugin_dispatch_values_wrapper(value_list_t const *vl) {
//   LOAD(plugin_dispatch_values);
//   return (*plugin_dispatch_values_ptr)(vl);
// }
//
// cdtime_t plugin_get_interval_wrapper(void) {
//   LOAD(plugin_get_interval);
//   return (*plugin_get_interval_ptr)();
// }
//
// int plugin_register_complex_read_wrapper(meta_data_t *group, char const *name,
//                                          plugin_read_cb callback,
//                                          cdtime_t interval, user_data_t *ud) {
//   LOAD(plugin_register_complex_read);
//   return (*plugin_register_complex_read_ptr)(group, name, callback, interval,
//                                              ud);
// }
//
// int plugin_register_log_wrapper(char const *name, plugin_log_cb callback,
//                                 user_data_t *ud) {
//   LOAD(plugin_register_log);
//   return (*plugin_register_log_ptr)(name, callback, ud);
// }
//
// int plugin_register_shutdown_wrapper(char const *name,
//                                      plugin_shutdown_cb callback) {
//   LOAD(plugin_register_shutdown);
//   return (*plugin_register_shutdown_ptr)(name, callback);
// }
//
// int plugin_register_write_wrapper(char const *name, plugin_write_cb callback,
//                                   user_data_t *ud) {
//   LOAD(plugin_register_write);
//   return (*plugin_register_write_ptr)(name, callback, ud);
// }
import "C"
