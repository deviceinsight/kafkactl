package acl

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type ResourceACLEntry struct {
	ResourceType string  `json:"resourceType" yaml:"resourceType"`
	ResourceName string  `json:"resourceName" yaml:"resourceName"`
	PatternType  string  `json:"patternType" yaml:"patternType"`
	Acls         []Entry `json:"acls" yaml:"acls"`
}

type Entry struct {
	Principal      string
	Host           string
	Operation      string
	PermissionType string `json:"permissionType" yaml:"permissionType"`
}

type GetACLFlags struct {
	OutputFormat string
	FilterTopic  string
	Operation    string
	ResourceName string
	PatternType  string
	Principal    string
	Host         string
	Allow        bool
	Deny         bool
	Topics       bool
	Groups       bool
	Cluster      bool
}

type CreateACLFlags struct {
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

type DeleteACLFlags struct {
	ValidateOnly bool
	Topics       bool
	Groups       bool
	Cluster      bool
	Allow        bool
	Deny         bool
	Principal    string
	Host         string
	Operation    string
	PatternType  string
}

type Operation struct {
}

func (operation *Operation) GetACL(flags GetACLFlags) error {

	var (
		ctx   internal.ClientContext
		err   error
		admin sarama.ClusterAdmin
		acls  []sarama.ResourceAcls
	)

	if ctx, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&ctx); err != nil {
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

	if flags.ResourceName != "" {
		filter.ResourceName = &flags.ResourceName
	}

	if flags.Principal != "" {
		filter.Principal = &flags.Principal
	}

	if flags.Host != "" {
		filter.Host = &flags.Host
	}

	if acls, err = admin.ListAcls(filter); err != nil {
		return errors.Wrap(err, "failed to list acls")
	}

	aclList := make([]ResourceACLEntry, 0)

	for _, acl := range acls {
		resourceACL := ResourceACLEntry{
			ResourceType: resourceTypeToString(acl.ResourceType),
			ResourceName: acl.ResourceName,
			PatternType:  patternTypeToString(acl.ResourcePatternType),
			Acls:         make([]Entry, 0),
		}

		for _, ac := range acl.Acls {
			resourceACL.Acls = append(resourceACL.Acls, Entry{
				Principal:      ac.Principal,
				Host:           ac.Host,
				Operation:      operationToString(ac.Operation),
				PermissionType: permissionTypeToString(ac.PermissionType),
			})
		}
		aclList = append(aclList, resourceACL)
	}

	return printResourceAcls(flags.OutputFormat, aclList...)
}

func (operation *Operation) CreateACL(flags CreateACLFlags) error {

	var (
		ctx   internal.ClientContext
		err   error
		admin sarama.ClusterAdmin
	)

	if ctx, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&ctx); err != nil {
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
		resource.ResourceName = "kafka-cluster"
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

	resourceACL := ResourceACLEntry{
		ResourceType: resourceTypeToString(resource.ResourceType),
		ResourceName: resource.ResourceName,
		PatternType:  patternTypeToString(resource.ResourcePatternType),
		Acls:         make([]Entry, 0),
	}

	for _, acl := range acls {

		if !flags.ValidateOnly {
			if err = admin.CreateACL(resource, acl); err != nil {
				return errors.Wrap(err, "failed to create acl")
			}
		}

		resourceACL.Acls = append(resourceACL.Acls, Entry{
			Principal:      acl.Principal,
			Host:           acl.Host,
			Operation:      operationToString(acl.Operation),
			PermissionType: permissionTypeToString(acl.PermissionType),
		})
	}

	return printResourceAcls("", resourceACL)
}

func (operation *Operation) DeleteACL(flags DeleteACLFlags) error {

	var (
		ctx         internal.ClientContext
		err         error
		admin       sarama.ClusterAdmin
		matchingACL []sarama.MatchingAcl
	)

	if ctx, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&ctx); err != nil {
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

	if flags.Principal != "" {
		filter.Principal = &flags.Principal
	}

	if flags.Host != "" {
		filter.Host = &flags.Host
	}

	if matchingACL, err = admin.DeleteACL(filter, flags.ValidateOnly); err != nil {
		return errors.Wrap(err, "failed to delete acl")
	}

	aclList := make([]ResourceACLEntry, 0)

	for _, acl := range matchingACL {
		resourceACL := ResourceACLEntry{
			ResourceType: resourceTypeToString(acl.ResourceType),
			ResourceName: acl.ResourceName,
			PatternType:  patternTypeToString(acl.ResourcePatternType),
			Acls:         make([]Entry, 1),
		}
		resourceACL.Acls[0].Principal = acl.Principal
		resourceACL.Acls[0].Host = acl.Host
		resourceACL.Acls[0].Operation = operationToString(acl.Operation)
		resourceACL.Acls[0].PermissionType = permissionTypeToString(acl.PermissionType)
		aclList = append(aclList, resourceACL)
	}

	return printResourceAcls("", aclList...)
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

func printResourceAcls(outputFormat string, aclList ...ResourceACLEntry) error {
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
		for _, resourceACL := range aclList {
			for _, aclEntry := range resourceACL.Acls {
				if err := tableWriter.Write(resourceACL.ResourceType, resourceACL.ResourceName, resourceACL.PatternType,
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

func CompleteCreateACL(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	output.Infof("complete")
	return nil, cobra.ShellCompDirectiveError
}

func FromYaml(yamlString string) ([]ResourceACLEntry, error) {
	var entries []ResourceACLEntry
	err := yaml.Unmarshal([]byte(yamlString), &entries)
	return entries, err
}
