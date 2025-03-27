package acl

import (
	"strings"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

func permissionTypeToString(permissionType sarama.AclPermissionType) string {
	switch permissionType {
	case sarama.AclPermissionUnknown:
		return "Unknown"
	case sarama.AclPermissionAny:
		return "Any"
	case sarama.AclPermissionDeny:
		return "Deny"
	case sarama.AclPermissionAllow:
		return "Allow"
	default:
		output.Warnf("unknown permissionType: %v", permissionType)
		return ""
	}
}

func permissionTypeFromString(permissionType string) sarama.AclPermissionType {
	switch strings.ToLower(permissionType) {
	case "unknown":
		return sarama.AclPermissionUnknown
	case "any":
		return sarama.AclPermissionAny
	case "deny":
		return sarama.AclPermissionDeny
	case "allow":
		return sarama.AclPermissionAllow
	default:
		output.Warnf("unknown permissionType: %v", permissionType)
		return sarama.AclPermissionUnknown
	}
}

func operationToString(operation sarama.AclOperation) string {
	switch operation {
	case sarama.AclOperationUnknown:
		return "Unknown"
	case sarama.AclOperationAny:
		return "Any"
	case sarama.AclOperationAll:
		return "All"
	case sarama.AclOperationRead:
		return "Read"
	case sarama.AclOperationWrite:
		return "Write"
	case sarama.AclOperationCreate:
		return "Create"
	case sarama.AclOperationDelete:
		return "Delete"
	case sarama.AclOperationAlter:
		return "Alter"
	case sarama.AclOperationDescribe:
		return "Describe"
	case sarama.AclOperationClusterAction:
		return "ClusterAction"
	case sarama.AclOperationDescribeConfigs:
		return "DescribeConfigs"
	case sarama.AclOperationAlterConfigs:
		return "AlterConfigs"
	case sarama.AclOperationIdempotentWrite:
		return "IdempotentWrite"
	default:
		output.Warnf("unknown operation: %v", operation)
		return ""
	}
}

func operationFromString(operation string) sarama.AclOperation {
	switch strings.ToLower(operation) {
	case "unknown":
		return sarama.AclOperationUnknown
	case "any":
		return sarama.AclOperationAny
	case "all":
		return sarama.AclOperationAll
	case "read":
		return sarama.AclOperationRead
	case "write":
		return sarama.AclOperationWrite
	case "create":
		return sarama.AclOperationCreate
	case "delete":
		return sarama.AclOperationDelete
	case "alter":
		return sarama.AclOperationAlter
	case "describe":
		return sarama.AclOperationDescribe
	case "clusteraction":
		return sarama.AclOperationClusterAction
	case "describeconfigs":
		return sarama.AclOperationDescribeConfigs
	case "alterconfigs":
		return sarama.AclOperationAlterConfigs
	case "idempotentwrite":
		return sarama.AclOperationIdempotentWrite
	default:
		output.Warnf("unknown operation: %v", operation)
		return sarama.AclOperationUnknown
	}
}

func patternTypeToString(patternType sarama.AclResourcePatternType) string {
	switch patternType {
	case sarama.AclPatternUnknown:
		return "Unknown"
	case sarama.AclPatternAny:
		return "Any"
	case sarama.AclPatternMatch:
		return "Match"
	case sarama.AclPatternLiteral:
		return "Literal"
	case sarama.AclPatternPrefixed:
		return "Prefixed"
	default:
		output.Warnf("unknown pattern type: %v", patternType)
		return ""
	}
}

func patternTypeFromString(patternType string) sarama.AclResourcePatternType {
	switch strings.ToLower(patternType) {
	case "unknown":
		return sarama.AclPatternUnknown
	case "any":
		return sarama.AclPatternAny
	case "match":
		return sarama.AclPatternMatch
	case "literal":
		return sarama.AclPatternLiteral
	case "prefixed":
		return sarama.AclPatternPrefixed
	default:
		output.Warnf("unknown pattern type: %v", patternType)
		return sarama.AclPatternUnknown
	}
}

func resourceTypeToString(resourceType sarama.AclResourceType) string {
	switch resourceType {
	case sarama.AclResourceUnknown:
		return "Unknown"
	case sarama.AclResourceAny:
		return "Any"
	case sarama.AclResourceTopic:
		return "Topic"
	case sarama.AclResourceGroup:
		return "Group"
	case sarama.AclResourceCluster:
		return "Cluster"
	case sarama.AclResourceTransactionalID:
		return "TransactionalID"
	default:
		output.Warnf("unknown resource type: %v", resourceType)
		return ""
	}
}

func resourceTypeFromString(resourceType string) sarama.AclResourceType {
	switch strings.ToLower(resourceType) {
	case "unknown":
		return sarama.AclResourceUnknown
	case "any":
		return sarama.AclResourceAny
	case "topic":
		return sarama.AclResourceTopic
	case "group":
		return sarama.AclResourceGroup
	case "cluster":
		return sarama.AclResourceCluster
	case "transactionalid":
		return sarama.AclResourceTransactionalID
	default:
		output.Warnf("unknown resource type: %v", resourceType)
		return sarama.AclResourceUnknown
	}
}
