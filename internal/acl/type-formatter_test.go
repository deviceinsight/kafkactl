package acl

import (
	"fmt"
	"testing"

	"github.com/IBM/sarama"
)

func TestOperationFromString(t *testing.T) {
	tests := []sarama.AclOperation{
		sarama.AclOperationUnknown,
		sarama.AclOperationAny,
		sarama.AclOperationAll,
		sarama.AclOperationRead,
		sarama.AclOperationWrite,
		sarama.AclOperationCreate,
		sarama.AclOperationDelete,
		sarama.AclOperationAlter,
		sarama.AclOperationDescribe,
		sarama.AclOperationClusterAction,
		sarama.AclOperationDescribeConfigs,
		sarama.AclOperationAlterConfigs,
		sarama.AclOperationIdempotentWrite,
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test operation %d", tt), func(t *testing.T) {
			stringOperation := operationToString(tt)
			if got := operationFromString(stringOperation); got != tt {
				t.Errorf("operationFromString() = %v, want %v", got, tt)
			}
		})
	}
}

func TestPatternTypeFromString(t *testing.T) {
	tests := []sarama.AclResourcePatternType{
		sarama.AclPatternUnknown,
		sarama.AclPatternAny,
		sarama.AclPatternMatch,
		sarama.AclPatternLiteral,
		sarama.AclPatternPrefixed,
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test operation %d", tt), func(t *testing.T) {
			stringOperation := patternTypeToString(tt)
			if got := patternTypeFromString(stringOperation); got != tt {
				t.Errorf("patternTypeFromString() = %v, want %v", got, tt)
			}
		})
	}
}

func TestPermissionTypeFromString(t *testing.T) {
	tests := []sarama.AclPermissionType{
		sarama.AclPermissionUnknown,
		sarama.AclPermissionAny,
		sarama.AclPermissionDeny,
		sarama.AclPermissionAllow,
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test operation %d", tt), func(t *testing.T) {
			stringOperation := permissionTypeToString(tt)
			if got := permissionTypeFromString(stringOperation); got != tt {
				t.Errorf("permissionTypeFromString() = %v, want %v", got, tt)
			}
		})
	}
}

func TestResourceTypeFromString(t *testing.T) {
	tests := []sarama.AclResourceType{
		sarama.AclResourceUnknown,
		sarama.AclResourceAny,
		sarama.AclResourceTopic,
		sarama.AclResourceGroup,
		sarama.AclResourceCluster,
		sarama.AclResourceTransactionalID,
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("test operation %d", tt), func(t *testing.T) {
			stringOperation := resourceTypeToString(tt)
			if got := resourceTypeFromString(stringOperation); got != tt {
				t.Errorf("resourceTypeFromString() = %v, want %v", got, tt)
			}
		})
	}
}
