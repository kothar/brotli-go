// Code shared between the enc and dec packages which would otherwise cause
// duplicate symbols during linking
package shared

/*
#import "dictionary.h"
*/
import "C"

import "unsafe"

func GetDictionary() unsafe.Pointer {
	return unsafe.Pointer(&C.sharedBrotliDictionary[0])
}
