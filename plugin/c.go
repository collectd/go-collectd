// +build go1.5,cgo

package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include "plugin.h"
// #include <stdlib.h>
// #include <dlfcn.h>
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
// static int *timeout_ptr;
// int timeout_wrapper(void) {
//   if (timeout_ptr == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     timeout_ptr = dlsym(hnd, "timeout_g");
//     dlclose(hnd);
//   }
//   return *timeout_ptr;
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
