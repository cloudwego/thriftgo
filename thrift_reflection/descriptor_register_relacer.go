// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package thrift_reflection

type FileDescriptorReplacer struct {
	RemoteGoPkgPath  string
	CurrentGoPkgPath string
	CurrentFilepath  string
	Matcher          string
}

func checkMatch(matcher string, f *FileDescriptor) bool {
	return f.GetStructDescriptor(matcher) != nil ||
		f.GetEnumDescriptor(matcher) != nil ||
		f.GetTypedefDescriptor(matcher) != nil ||
		f.GetExceptionDescriptor(matcher) != nil ||
		f.GetUnionDescriptor(matcher) != nil ||
		f.GetConstDescriptor(matcher) != nil ||
		f.GetServiceDescriptor(matcher) != nil
}

func ReplaceFileDescriptor(replacer *FileDescriptorReplacer) *FileDescriptor {
	remoteGoPkgPath := replacer.RemoteGoPkgPath
	currentGoPkgPath := replacer.CurrentGoPkgPath
	currentFilepath := replacer.CurrentFilepath
	matcher := replacer.Matcher
	remoteDesc := matchRemoteFileDescriptor(remoteGoPkgPath, matcher)
	if remoteDesc == nil {
		panic("not found remote fd")
	}
	var shadowDesc *FileDescriptor

	// make a copy from remoteDesc and replace metadata such as filepath、go pkg path to local's.
	shadowDesc = new(FileDescriptor)
	*shadowDesc = *remoteDesc
	shadowDesc.Filepath = currentFilepath
	shadowDesc.Extra = nil
	shadowDesc.setGoPkgPathRef(currentGoPkgPath)

	if remoteDesc.Filepath != currentFilepath {
		checkDuplicateAndRegister(shadowDesc, currentGoPkgPath)
	}
	return shadowDesc
}

func matchRemoteFileDescriptor(remoteGoPkgPath, matcher string) *FileDescriptor {
	for k, fd := range globalFD {
		if fd.checkGoPkgPathWithRef(remoteGoPkgPath) && checkMatch(matcher, fd) {
			return globalFD[k]
		}
	}
	return nil
}

func (f *FileDescriptor) setGoPkgPathRef(local string) {
	if f.Extra == nil {
		f.Extra = map[string]string{}
	}
	f.Extra["GoPkgPathRef"] = f.Extra["GoPkgPath"]
	f.Extra["GoPkgPath"] = local
}

func (f *FileDescriptor) checkGoPkgPathWithRef(path string) bool {
	if f.Extra == nil {
		return false
	}
	return f.Extra["GoPkgPath"] == path || f.Extra["GoPkgPathRef"] == path
}
