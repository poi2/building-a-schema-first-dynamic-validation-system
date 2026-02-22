package validator

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// CreateTestDescriptorBytes creates a test FileDescriptorSet from the user.v1 package
// This function is exported for use in other test packages
func CreateTestDescriptorBytes(t *testing.T) []byte {
	t.Helper()

	// Get file descriptor for user.v1 package
	fileDesc, err := protoregistry.GlobalFiles.FindFileByPath("user/v1/user.proto")
	if err != nil {
		t.Fatalf("failed to find user.proto: %v", err)
	}

	// Create FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{}

	// Add all dependencies recursively
	visited := make(map[string]bool)
	var addFile func(fd protoreflect.FileDescriptor)
	addFile = func(fd protoreflect.FileDescriptor) {
		path := fd.Path()
		if visited[path] {
			return
		}
		visited[path] = true

		// Add dependencies first
		for i := 0; i < fd.Imports().Len(); i++ {
			addFile(fd.Imports().Get(i).FileDescriptor)
		}

		// Add this file
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fd))
	}

	addFile(fileDesc)

	// Marshal to bytes
	data, err := proto.Marshal(fds)
	if err != nil {
		t.Fatalf("failed to marshal descriptor set: %v", err)
	}

	return data
}
