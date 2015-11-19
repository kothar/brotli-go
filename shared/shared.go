// Code shared between the enc and dec packages
package shared

/*
#include "dictionary.h"
*/
import "C"

import "unsafe"

func GetDictionary() unsafe.Pointer {
	return unsafe.Pointer(&C.sharedBrotliDictionary)
}
