//go:build darwin && arm64 && cgo && opengltest

package opengl

/*
#cgo darwin LDFLAGS: -framework OpenGL
#include <OpenGL/OpenGL.h>
#include <OpenGL/CGLCurrent.h>
#include <OpenGL/gl3.h>

static CGLError r289ChooseUploadFormat(CGLPixelFormatObj *format, GLint *count) {
    const CGLPixelFormatAttribute attributes[] = {
        kCGLPFAOpenGLProfile, (CGLPixelFormatAttribute)kCGLOGLPVersion_GL4_Core,
        kCGLPFAAccelerated,
        kCGLPFAColorSize, (CGLPixelFormatAttribute)24,
        kCGLPFAAlphaSize, (CGLPixelFormatAttribute)8,
        (CGLPixelFormatAttribute)0
    };
    return CGLChoosePixelFormat(attributes, format, count);
}

static GLint r289InitialUnpackAlignment(void) {
    GLint value = 0;
    glGetIntegerv(GL_UNPACK_ALIGNMENT, &value);
    return value;
}
*/
import "C"

import (
	"errors"
	"fmt"
)

// This explicitly tagged test bridge owns a windowless context, not a window,
// drawable, SDL subsystem or borrowed context. Tests call it on a locked thread.
type uploadContextR289 struct {
	format        C.CGLPixelFormatObj
	context       C.CGLContextObj
	initialUnpack int32
}

func cglUploadErrorR289(operation string, code C.CGLError) error {
	if code == C.kCGLNoError {
		return nil
	}
	return fmt.Errorf("%s: CGLError=%d %s", operation, int(code), C.GoString(C.CGLErrorString(code)))
}

func newUploadContextR289() (*uploadContextR289, error) {
	if C.CGLGetCurrentContext() != nil {
		return nil, errors.New("R289 requires a fresh thread with no current CGL context")
	}
	owner := &uploadContextR289{}
	var count C.GLint
	if err := cglUploadErrorR289("CGLChoosePixelFormat", C.r289ChooseUploadFormat(&owner.format, &count)); err != nil {
		return nil, errors.Join(err, owner.close())
	}
	if owner.format == nil || count < 1 {
		return nil, errors.Join(errors.New("CGL returned no owned pixel format"), owner.close())
	}
	if err := cglUploadErrorR289("CGLCreateContext", C.CGLCreateContext(owner.format, nil, &owner.context)); err != nil {
		return nil, errors.Join(err, owner.close())
	}
	if owner.context == nil {
		return nil, errors.Join(errors.New("CGL returned no owned context"), owner.close())
	}
	if err := cglUploadErrorR289("CGLSetCurrentContext", C.CGLSetCurrentContext(owner.context)); err != nil {
		return nil, errors.Join(err, owner.close())
	}
	owner.initialUnpack = int32(C.r289InitialUnpackAlignment())
	if code := C.glGetError(); code != C.GL_NO_ERROR {
		return nil, errors.Join(fmt.Errorf("initial unpack query: GLError=%#x", uint32(code)), owner.close())
	}
	return owner, nil
}

func (owner *uploadContextR289) close() error {
	var failures []error
	if owner.context != nil {
		failures = append(failures, cglUploadErrorR289("CGLSetCurrentContext(nil)", C.CGLSetCurrentContext(nil)))
		if C.CGLGetCurrentContext() != nil {
			failures = append(failures, errors.New("CGL context remained current after unbinding"))
		}
		failures = append(failures, cglUploadErrorR289("CGLDestroyContext", C.CGLDestroyContext(owner.context)))
		owner.context = nil
	}
	if owner.format != nil {
		failures = append(failures, cglUploadErrorR289("CGLDestroyPixelFormat", C.CGLDestroyPixelFormat(owner.format)))
		owner.format = nil
	}
	return errors.Join(failures...)
}
