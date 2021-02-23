package acl

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ResourceAclEntry struct {
	ResourceType string     `json:"resourceType" yaml:"resourceType"`
	ResourceName string     `json:"resourceName" yaml:"resourceName"`
	PatternType  string     `json:"patternType" yaml:"patternType"`
	Acls         []AclEntry `json:"acls" yaml:"acls"`
}

type AclEntry struct {
	Principal      string
	Host           string
	Operation      string
	PermissionType string `json:"permissionType" yaml:"permissionType"`
}

type GetAclFlags struct {
	OutputFormat string
	FilterTopic  string
	Operation    string
	PatternType  string
	Allow        bool
	Deny         bool
	Topics       bool
	Groups       bool
	Cluster      bool
}

type CreateAclFlags struct {
	Principal    string
	Hosts        []string
	Operations   []string
	Allow        bool
	Deny         bool
	Topic        string
	Group        string
	Cluster      bool
	PatternType  string
	ValidateOnly bool
}

type DeleteAclFlags struct {
	ValidateOnly bool
	Topics       bool
	Groups       bool
	Cluster      bool
	Allow        bool
	Deny         bool
	Operation    string
	PatternType  string
}

type AclOperation struct {
}

func (operation *AclOperation) GetAcl(flags GetAclFlags) error {

	var (
		ctx   operations.ClientContext
		err   error
		admin sarama.ClusterAdmin
		acls  []sarama.ResourceAcls
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if flags.Operation == "" {
		output.Debugf("defaulting patternType to: any")
		flags.Operation = "any"
	}

	if flags.PatternType == "" {
		output.Debugf("defaulting patternType to: any")
		flags.PatternType = "any"
	}

	if flags.Allow && flags.Deny {
		return errors.New("--allow and --deny cannot be provided both")
	}

	permissionType := sarama.AclPermissionAny
	if flags.Deny {
		permissionType = sarama.AclPermissionDeny
	} else if flags.Allow {
		permissionType = sarama.AclPermissionAllow
	}

	filter := sarama.AclFilter{
		PermissionType: permissionType,
		Operation:      operationFromString(flags.Operation),
	}

	if flags.Topics {
		filter.ResourceType = sarama.AclResourceTopic
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	} else if flags.Groups {
		filter.ResourceType = sarama.AclResourceGroup
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	} else if flags.Cluster {
		filter.ResourceType = sarama.AclResourceCluster
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	} else {
		filter.ResourceType = sarama.AclResourceAny
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	}

	if acls, err = admin.ListAcls(filter); err != nil {
		return errors.Wrap(err, "failed to list acls")
	}

	aclList := make([]ResourceAclEntry, 0)

	for _, acl := range acls {
		resourceAcl := ResourceAclEntry{
			ResourceType: resourceTypeToString(acl.ResourceType),
			ResourceName: acl.ResourceName,
			PatternType:  patternTypeToString(acl.ResourcePatternType),
			Acls:         make([]AclEntry, 0),
		}

		for _, ac := range acl.Acls {
			resourceAcl.Acls = append(resourceAcl.Acls, AclEntry{
				Principal:      ac.Principal,
				Host:           ac.Host,
				Operation:      operationToString(ac.Operation),
				PermissionType: permissionTypeToString(ac.PermissionType),
			})
		}
		aclList = append(aclList, resourceAcl)
	}

	if err = printResourceAcls(flags.OutputFormat, aclList...); err != nil {
		return err
	} else {
		return nil
	}
}

func (operation *AclOperation) CreateAcl(flags CreateAclFlags) error {

	var (
		ctx   operations.ClientContext
		err   error
		admin sarama.ClusterAdmin
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if flags.Principal == "" {
		return errors.New("principal must be set")
	}

	if len(flags.Hosts) == 0 {
		output.Debugf("no host specified. Using *")
		flags.Hosts = append(flags.Hosts, "*")
	}

	if len(flags.Operations) == 0 {
		return errors.New("at least one operation has to be specified")
	}

	if !xor(flags.Allow, flags.Deny) {
		return errors.New("either --allow or --deny has to be provided")
	}

	if !xor(flags.Topic != "", flags.Group != "", flags.Cluster) {
		return errors.New("either --topic=topic-name or --group=group-name or --cluster has to be provided")
	}

	resource := sarama.Resource{}

	if flags.Topic != "" {
		resource.ResourceType = sarama.AclResourceTopic
		resource.ResourceName = flags.Topic
		resource.ResourcePatternType = patternTypeFromString(flags.PatternType)
	} else if flags.Group != "" {
		resource.ResourceType = sarama.AclResourceGroup
		resource.ResourceName = flags.Group
		resource.ResourcePatternType = patternTypeFromString(flags.PatternType)
	} else {
		resource.ResourceType = sarama.AclResourceCluster
		resource.ResourcePatternType = patternTypeFromString(flags.PatternType)
	}

	permissionType := sarama.AclPermissionAllow
	if flags.Deny {
		permissionType = sarama.AclPermissionDeny
	}

	acls := make([]sarama.Acl, 0)

	for _, host := range flags.Hosts {
		for _, operation := range flags.Operations {
			acl := sarama.Acl{
				Principal:      flags.Principal,
				Host:           host,
				Operation:      operationFromString(operation),
				PermissionType: permissionType,
			}
			acls = append(acls, acl)
		}
	}

	resourceAcl := ResourceAclEntry{
		ResourceType: resourceTypeToString(resource.ResourceType),
		ResourceName: resource.ResourceName,
		PatternType:  patternTypeToString(resource.ResourcePatternType),
		Acls:         make([]AclEntry, 0),
	}

	for _, acl := range acls {

		if !flags.ValidateOnly {
			if err = admin.CreateACL(resource, acl); err != nil {
				return errors.Wrap(err, "failed to create acl")
			}
		}

		resourceAcl.Acls = append(resourceAcl.Acls, AclEntry{
			Principal:      acl.Principal,
			Host:           acl.Host,
			Operation:      operationToString(acl.Operation),
			PermissionType: permissionTypeToString(acl.PermissionType),
		})
	}

	if err = printResourceAcls("", resourceAcl); err != nil {
		return err
	} else {
		return nil
	}
}

func (operation *AclOperation) DeleteAcl(flags DeleteAclFlags) error {

	var (
		ctx         operations.ClientContext
		err         error
		admin       sarama.ClusterAdmin
		matchingAcl []sarama.MatchingAcl
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if flags.Operation == "" {
		return errors.New("no operation has been specified")
	}

	if flags.PatternType == "" {
		return errors.New("no pattern has been specified")
	}

	if flags.Allow && flags.Deny {
		return errors.New("--allow and --deny cannot be provided both")
	}

	if !xor(flags.Topics, flags.Groups, flags.Cluster) {
		return errors.New("either --topic or --group or --cluster has to be provided")
	}

	permissionType := sarama.AclPermissionAny
	if flags.Deny {
		permissionType = sarama.AclPermissionDeny
	} else if flags.Allow {
		permissionType = sarama.AclPermissionAllow
	}

	filter := sarama.AclFilter{
		PermissionType: permissionType,
		Operation:      operationFromString(flags.Operation),
	}

	if flags.Topics {
		filter.ResourceType = sarama.AclResourceTopic
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	} else if flags.Groups {
		filter.ResourceType = sarama.AclResourceGroup
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	} else {
		filter.ResourceType = sarama.AclResourceCluster
		filter.ResourcePatternTypeFilter = patternTypeFromString(flags.PatternType)
	}

	if matchingAcl, err = admin.DeleteACL(filter, flags.ValidateOnly); err != nil {
		return errors.Wrap(err, "failed to delete acl")
	}

	aclList := make([]ResourceAclEntry, 0)

	for _, acl := range matchingAcl {
		resourceAcl := ResourceAclEntry{
			ResourceType: resourceTypeToString(acl.ResourceType),
			ResourceName: acl.ResourceName,
			PatternType:  patternTypeToString(acl.ResourcePatternType),
			Acls:         make([]AclEntry, 1),
		}
		resourceAcl.Acls[0].Principal = acl.Principal
		resourceAcl.Acls[0].Host = acl.Host
		resourceAcl.Acls[0].Operation = operationToString(acl.Operation)
		resourceAcl.Acls[0].PermissionType = permissionTypeToString(acl.PermissionType)
		aclList = append(aclList, resourceAcl)
	}

	if err = printResourceAcls("", aclList...); err != nil {
		return err
	} else {
		return nil
	}
}

func xor(values ...bool) bool {
	and := true
	or := false

	for _, value := range values {
		and = and && value
		or = or || value
	}

	return or && !and
}

func printResourceAcls(outputFormat string, aclList ...ResourceAclEntry) error {
	tableWriter := output.CreateTableWriter()
	if outputFormat == "" {
		if err := tableWriter.WriteHeader("RESOURCE_TYPE", "RESOURCE_NAME", "PATTERN_TYPE", "PRINCIPAL", "HOST", "OPERATION", "PERMISSION_TYPE"); err != nil {
			return err
		}
	} else if outputFormat == "json" || outputFormat == "yaml" {
		if err := output.PrintObject(aclList, outputFormat); err != nil {
			return err
		}
	} else {
		return errors.Errorf("unknown output format: %s", outputFormat)
	}

	if outputFormat == "" {
		for _, resourceAcl := range aclList {
			for _, aclEntry := range resourceAcl.Acls {
				if err := tableWriter.Write(resourceAcl.ResourceType, resourceAcl.ResourceName, resourceAcl.PatternType,
					aclEntry.Principal, aclEntry.Host, aclEntry.Operation, aclEntry.PermissionType); err != nil {
					return err
				}
			}
		}

		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func CompleteCreateAcl(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	output.Infof("complete")
	return nil, cobra.ShellCompDirectiveError
}
