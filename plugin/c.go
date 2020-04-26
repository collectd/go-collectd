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
// static cdtime_t (*plugin_get_interval_ptr)(void);
// cdtime_t plugin_get_interval_wrapper(void) {
//   if (plugin_get_interval_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     plugin_get_interval_ptr = dlsym(hnd, "plugin_get_interval");
//     dlclose(hnd);
//   }
//   return (*plugin_get_interval_ptr) ();
// }
//
// void value_list_add (value_list_t *vl, value_t v) {
//   value_t *tmp = realloc (vl->values, sizeof(v) * (vl->values_len + 1));
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
// int (*register_log_ptr) (char const *, plugin_log_cb, user_data_t const *);
// int register_log_wrapper (char const *name, plugin_log_cb callback, user_data_t const *user_data) {
//   if (register_log_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_log_ptr = dlsym(hnd, "plugin_register_log");
//     dlclose(hnd);
//   }
//   return (*register_log_ptr) (name, callback, user_data);
// }
//
// typedef int (*plugin_complex_config_cb)(oconfig_item_t *);
//
// static int (*register_complex_config_ptr) (const char *, plugin_complex_config_cb);
// int register_complex_config_wrapper (const char *name, plugin_complex_config_cb callback) {
//   if (register_complex_config_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_complex_config_ptr = dlsym(hnd, "plugin_register_complex_config");
//     dlclose(hnd);
//   }
//   return (*register_complex_config_ptr) (name, callback);
// }
//
// static int (*register_init_ptr) (const char *, plugin_init_cb);
// int register_init_wrapper (const char *name, plugin_init_cb callback) {
//   if (register_init_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     register_init_ptr = dlsym(hnd, "plugin_register_init");
//     dlclose(hnd);
//   }
//   return (*register_init_ptr) (name, callback);
// }
import "C"
