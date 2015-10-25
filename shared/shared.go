// Code shared between the enc and dec packages which would otherwise cause
// duplicate symbols during linking
package shared

/*
#import "dictionary.h"
*/
import "C"

import "unsafe"

func GetDictionary() []byte {
	return C.GoBytes(unsafe.Pointer(&C.sharedBrotliDictionary[0]), 122784)
}
