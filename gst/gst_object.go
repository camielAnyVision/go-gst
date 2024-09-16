package gst

import "C"
import (
	"github.com/tinyzimmer/go-glib/glib"
	"unsafe"
)

/*
#include "gst.go.h"
#cgo CFLAGS: -I /usr/include/glib-2.0/
#cgo pkg-config: glib-2.0 gobject-2.0
#include "glib-object.h"

int refcount(void* x) {
   GObject* obj = ((GObject*)x);
   return obj->ref_count;
}
*/
import "C"

import (
	"time"
)

// Object is a go representation of a GstObject.
type Object struct{ *glib.InitiallyUnowned }

// FromGstObjectUnsafeNone returns an Object wrapping the given pointer. It meant for internal
// usage and exported for visibility to other packages.
func FromGstObjectUnsafeNone(ptr unsafe.Pointer) *Object { return wrapObject(glib.TransferNone(ptr)) }

// FromGstObjectUnsafeFull returns an Object wrapping the given pointer. It meant for internal
// usage and exported for visibility to other packages.
func FromGstObjectUnsafeFull(ptr unsafe.Pointer) *Object { return wrapObject(glib.TransferFull(ptr)) }

// Instance returns the native C GstObject.
func (o *Object) Instance() *C.GstObject { return C.toGstObject(o.Unsafe()) }

// BaseObject is a convenience method for retrieving this object from embedded structs.
func (o *Object) BaseObject() *Object { return o }

// GstObject is an alias to Instance on the underlying GstObject of any extending struct.
func (o *Object) GstObject() *C.GstObject { return C.toGstObject(o.Unsafe()) }

// GObject returns the underlying GObject instance.
func (o *Object) GObject() *glib.Object { return o.InitiallyUnowned.Object }

// GetName returns the name of this object.
func (o *Object) GetName() string {
	cName := C.gst_object_get_name((*C.GstObject)(o.Instance()))
	defer C.free(unsafe.Pointer(cName))
	return C.GoString(cName)
}

// GetValue retrieves the value for the given controlled property at the given timestamp.
func (o *Object) GetValue(property string, timestamp time.Duration) *glib.Value {
	cprop := C.CString(property)
	defer C.free(unsafe.Pointer(cprop))
	gval := C.gst_object_get_value(o.Instance(), (*C.gchar)(cprop), C.GstClockTime(timestamp.Nanoseconds()))
	if gval == nil {
		return nil
	}
	return glib.ValueFromNative(unsafe.Pointer(gval))
}

// SetArg sets the argument name to value on this object. Note that function silently returns
// if object has no property named name or when value cannot be converted to the type for this
// property.
func (o *Object) SetArg(name, value string) {
	cName := C.CString(name)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	C.gst_util_set_object_arg(
		(*C.GObject)(o.Unsafe()),
		(*C.gchar)(unsafe.Pointer(cName)),
		(*C.gchar)(unsafe.Pointer(cValue)),
	)
}

// Log logs a message to the given category from this object using the currently registered
// debugging handlers.
func (o *Object) Log(cat *DebugCategory, level DebugLevel, message string) {
	cat.logDepth(level, message, 2, (*C.GObject)(o.Unsafe()))
}

// Clear will will clear all references to this object. If the reference is already null
// the the function does nothing. Otherwise the reference count is decreased and the pointer
// set to null.
func (o *Object) Clear() {
	if ptr := o.Unsafe(); ptr != nil {
		C.gst_clear_object((**C.GstObject)(unsafe.Pointer(&ptr)))
	}
}

// Ref increments the reference count on object. This function does not take the lock on object
// because it relies on atomic refcounting. For convenience the same object is returned.
func (o *Object) Ref() *Object {
	C.gst_object_ref((C.gpointer)(o.Unsafe()))
	return o
}

func (o *Object) GetCurrRefcount() int {
	if o.InitiallyUnowned.Object.GObject == nil || uintptr(unsafe.Pointer(o.InitiallyUnowned.Object.GObject)) == 0 {
		return 0
	}
	return int(C.refcount(unsafe.Pointer(o.InitiallyUnowned.Object.GObject)))
}

// Unref decrements the reference count on object. If reference count hits zero, destroy object.
// This function does not take the lock on object as it relies on atomic refcounting.
func (o *Object) Unref() {
	if o.Object.GObject == nil || uintptr(unsafe.Pointer(o.InitiallyUnowned.Object.GObject)) == 0 {
		return
	}
	if o.GetCurrRefcount() != 0 {
		C.gst_object_unref((C.gpointer)(o.Unsafe()))
	}
}

func (o *Object) Terminate() {
	for i := 0; i < o.GetCurrRefcount(); i++ {
		if o.GetCurrRefcount() == 1 {
			break
		}
		o.Unref()
	}
}
