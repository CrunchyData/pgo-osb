package broker

/*
 Copyright 2017-2018 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

import (
	"encoding/base32"
	"encoding/hex"
)

// CompactUUIDString reduces the string representation of a UUID into a
// shortened base32 representation of the same bits
// Example Input:  "a7cb6bd8-cf67-400f-805c-019e85eac3bf"
// Example Output: "U7FWXWGPM5AA7AC4AGPIL2WDX4"
func CompactUUIDString(uuid string) (string, error) {
	buf := make([]byte, 0, len(uuid))
	for _, r := range uuid {
		if r != '-' {
			buf = append(buf, byte(r))
		}
	}

	unhex := make([]byte, hex.DecodedLen(len(buf)))
	_, err := hex.Decode(unhex, buf)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(unhex), nil
}
