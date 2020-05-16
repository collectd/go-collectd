package plugin

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <stdbool.h>
// #include <errno.h>
// #include "plugin.h"
//
// /* work-around because Go can't deal with fields named "type". */
// static int config_value_type(oconfig_value_t *v) {
//   if (v == NULL) {
//     errno = EINVAL;
//     return -1;
//   }
//   return v->type;
// }
//
// /* work-around because CGo has trouble accessing unions. */
// static char *config_value_string(oconfig_value_t *v) {
//   if (v == NULL || v->type != OCONFIG_TYPE_STRING) {
//     errno = EINVAL;
//     return NULL;
//   }
//   return v->value.string;
// }
// static double config_value_number(oconfig_value_t *v) {
//   if (v == NULL || v->type != OCONFIG_TYPE_NUMBER) {
//     errno = EINVAL;
//     return NAN;
//   }
//   return v->value.number;
// }
// static bool config_value_boolean(oconfig_value_t *v) {
//   if (v == NULL || v->type != OCONFIG_TYPE_BOOLEAN) {
//     errno = EINVAL;
//     return 0;
//   }
//   return v->value.boolean;
// }
import "C"

import (
	"fmt"
	"unsafe"

	"collectd.org/api"
)

func unmarshalConfigBlocks(blocks *C.oconfig_item_t, blocks_num C.int) ([]api.Config, error) {
	var ret []api.Config
	for i := C.int(0); i < blocks_num; i++ {
		offset := uintptr(i) * C.sizeof_oconfig_item_t
		cBlock := (*C.oconfig_item_t)(unsafe.Pointer(uintptr(unsafe.Pointer(blocks)) + offset))

		goBlock, err := unmarshalConfigBlock(cBlock)
		if err != nil {
			return nil, err
		}
		ret = append(ret, goBlock)
	}
	return ret, nil
}

func unmarshalConfigBlock(block *C.oconfig_item_t) (api.Config, error) {
	cfg := api.Config{
		Key: C.GoString(block.key),
	}

	var err error
	if cfg.Values, err = unmarshalConfigValues(block.values, block.values_num); err != nil {
		return api.Config{}, err
	}

	if cfg.Children, err = unmarshalConfigBlocks(block.children, block.children_num); err != nil {
		return api.Config{}, err
	}

	return cfg, nil
}

func unmarshalConfigValues(values *C.oconfig_value_t, values_num C.int) ([]api.ConfigValue, error) {
	var ret []api.ConfigValue
	for i := C.int(0); i < values_num; i++ {
		offset := uintptr(i) * C.sizeof_oconfig_value_t
		cValue := (*C.oconfig_value_t)(unsafe.Pointer(uintptr(unsafe.Pointer(values)) + offset))

		goValue, err := unmarshalConfigValue(cValue)
		if err != nil {
			return nil, err
		}
		ret = append(ret, goValue)
	}
	return ret, nil
}

func unmarshalConfigValue(value *C.oconfig_value_t) (api.ConfigValue, error) {
	typ, err := C.config_value_type(value)
	if err := wrapCError(0, err, "config_value_type"); err != nil {
		return api.ConfigValue{}, err
	}

	switch typ {
	case C.OCONFIG_TYPE_STRING:
		s, err := C.config_value_string(value)
		if err := wrapCError(0, err, "config_value_string"); err != nil {
			return api.ConfigValue{}, err
		}
		return api.ConfigString(C.GoString(s)), nil
	case C.OCONFIG_TYPE_NUMBER:
		n, err := C.config_value_number(value)
		if err := wrapCError(0, err, "config_value_number"); err != nil {
			return api.ConfigValue{}, err
		}
		return api.ConfigFloat64(float64(n)), nil
	case C.OCONFIG_TYPE_BOOLEAN:
		b, err := C.config_value_boolean(value)
		if err := wrapCError(0, err, "config_value_boolean"); err != nil {
			return api.ConfigValue{}, err
		}
		return api.ConfigBool(bool(b)), nil
	default:
		return api.ConfigValue{}, fmt.Errorf("unknown config value type: %d", typ)
	}
}
