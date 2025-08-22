package user

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
)

type User struct {
	Name       string                `json:"name" yaml:"name"`
	Mechanisms []ScramCredentialInfo `json:"mechanisms,omitempty" yaml:"mechanisms,omitempty"`
}

type ScramCredentialInfo struct {
	Mechanism  string `json:"mechanism" yaml:"mechanism"`
	Iterations int32  `json:"iterations" yaml:"iterations"`
}

type CreateUserFlags struct {
	Mechanism  string
	Password   string
	Salt       string
	Iterations int32
}

type AlterUserFlags struct {
	Mechanism  string
	Password   string
	Salt       string
	Iterations int32
}

type DeleteUserFlags struct {
	Mechanism string
}

type GetUsersFlags struct {
	OutputFormat string
}

type DescribeUserFlags struct {
	OutputFormat string
}

type Operation struct{}

func (operation *Operation) CreateUser(username string, flags CreateUserFlags) error {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}
	defer admin.Close()

	// Generate salt if not provided
	salt := []byte(flags.Salt)
	if len(salt) == 0 {
		salt, err = generateRandomSalt()
		if err != nil {
			return errors.Wrap(err, "failed to generate salt")
		}
	} else {
		// Decode base64 salt if provided
		salt, err = base64.StdEncoding.DecodeString(flags.Salt)
		if err != nil {
			return errors.Wrap(err, "failed to decode salt")
		}
	}

	// Convert mechanism string to Sarama type
	mechanism, err := mechanismFromString(flags.Mechanism)
	if err != nil {
		return err
	}

	// Set default iterations if not specified
	iterations := flags.Iterations
	if iterations <= 0 {
		iterations = 4096
	}

	// Create upsert request
	upsert := sarama.AlterUserScramCredentialsUpsert{
		Name:       username,
		Mechanism:  mechanism,
		Iterations: iterations,
		Salt:       salt,
		Password:   []byte(flags.Password),
	}

	// Use UpsertUserScramCredentials method
	response, err := admin.UpsertUserScramCredentials([]sarama.AlterUserScramCredentialsUpsert{upsert})
	if err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	// Check for errors in response
	for _, result := range response {
		if result.User == username && result.ErrorCode != sarama.ErrNoError {
			errorMsg := ""
			if result.ErrorMessage != nil {
				errorMsg = *result.ErrorMessage
			}
			return errors.Errorf("failed to create user '%s': %s", username, errorMsg)
		}
	}

	output.Infof("user '%s' has been created with %s", username, flags.Mechanism)
	return nil
}

func (operation *Operation) AlterUser(username string, flags AlterUserFlags) error {
	// Same logic as CreateUser - SCRAM credentials are upserted
	createFlags := CreateUserFlags(flags)

	err := operation.CreateUser(username, createFlags)
	if err != nil {
		return err
	}

	output.Infof("user '%s' credentials have been updated", username)
	return nil
}

func (operation *Operation) DeleteUser(username string, flags DeleteUserFlags) error {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}
	defer admin.Close()

	// Convert mechanism string to Sarama type
	mechanism, err := mechanismFromString(flags.Mechanism)
	if err != nil {
		return err
	}

	// Create delete request
	deleteReq := sarama.AlterUserScramCredentialsDelete{
		Name:      username,
		Mechanism: mechanism,
	}

	// Use DeleteUserScramCredentials method
	response, err := admin.DeleteUserScramCredentials([]sarama.AlterUserScramCredentialsDelete{deleteReq})
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	// Check for errors in response
	for _, result := range response {
		if result.User == username && result.ErrorCode != sarama.ErrNoError {
			errorMsg := ""
			if result.ErrorMessage != nil {
				errorMsg = *result.ErrorMessage
			}
			return errors.Errorf("failed to delete user '%s': %s", username, errorMsg)
		}
	}

	output.Infof("user '%s' %s credentials have been deleted", username, flags.Mechanism)
	return nil
}

func (operation *Operation) GetUsers(flags GetUsersFlags) error {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}
	defer admin.Close()

	// Get all users (empty list means all users)
	response, err := admin.DescribeUserScramCredentials([]string{})
	if err != nil {
		return errors.Wrap(err, "failed to get users")
	}

	// Convert response to our User format
	var users []User
	for _, result := range response {
		if result.ErrorCode != sarama.ErrNoError {
			errorMsg := ""
			if result.ErrorMessage != nil {
				errorMsg = *result.ErrorMessage
			}
			output.Warnf("error getting user '%s': %s", result.User, errorMsg)
			continue
		}

		user := User{
			Name:       result.User,
			Mechanisms: make([]ScramCredentialInfo, 0, len(result.CredentialInfos)),
		}

		for _, cred := range result.CredentialInfos {
			user.Mechanisms = append(user.Mechanisms, ScramCredentialInfo{
				Mechanism:  mechanismToString(cred.Mechanism),
				Iterations: cred.Iterations,
			})
		}

		// Sort mechanisms for consistent output
		sort.Slice(user.Mechanisms, func(i, j int) bool {
			return user.Mechanisms[i].Mechanism < user.Mechanisms[j].Mechanism
		})

		users = append(users, user)
	}

	// Sort users by name
	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(users, flags.OutputFormat)
	} else if flags.OutputFormat != "" {
		return errors.Errorf("unknown output format: %s", flags.OutputFormat)
	}

	// Default table format
	tableWriter := output.CreateTableWriter()
	if err := tableWriter.WriteHeader("USERNAME", "MECHANISMS"); err != nil {
		return err
	}

	for _, user := range users {
		mechanisms := make([]string, len(user.Mechanisms))
		for i, mech := range user.Mechanisms {
			mechanisms[i] = mech.Mechanism
		}
		if err := tableWriter.Write(user.Name, strings.Join(mechanisms, ",")); err != nil {
			return err
		}
	}

	return tableWriter.Flush()
}

func (operation *Operation) DescribeUser(username string, flags DescribeUserFlags) error {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}
	defer admin.Close()

	// Describe specific user
	response, err := admin.DescribeUserScramCredentials([]string{username})
	if err != nil {
		return errors.Wrap(err, "failed to describe user")
	}

	// Find the user in response
	for _, result := range response {
		if result.User == username {
			if result.ErrorCode != sarama.ErrNoError {
				errorMsg := ""
				if result.ErrorMessage != nil {
					errorMsg = *result.ErrorMessage
				}
				return errors.Errorf("user '%s' not found or error: %s", username, errorMsg)
			}

			user := User{
				Name:       result.User,
				Mechanisms: make([]ScramCredentialInfo, 0, len(result.CredentialInfos)),
			}

			for _, cred := range result.CredentialInfos {
				user.Mechanisms = append(user.Mechanisms, ScramCredentialInfo{
					Mechanism:  mechanismToString(cred.Mechanism),
					Iterations: cred.Iterations,
				})
			}

			// Sort mechanisms for consistent output
			sort.Slice(user.Mechanisms, func(i, j int) bool {
				return user.Mechanisms[i].Mechanism < user.Mechanisms[j].Mechanism
			})

			if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
				return output.PrintObject(user, flags.OutputFormat)
			} else if flags.OutputFormat != "" {
				return errors.Errorf("unknown output format: %s", flags.OutputFormat)
			}

			// Default table format
			tableWriter := output.CreateTableWriter()
			if err := tableWriter.WriteHeader("USERNAME", "MECHANISM", "ITERATIONS"); err != nil {
				return err
			}

			for _, mech := range user.Mechanisms {
				if err := tableWriter.Write(user.Name, mech.Mechanism, fmt.Sprint(mech.Iterations)); err != nil {
					return err
				}
			}

			return tableWriter.Flush()
		}
	}

	return errors.Errorf("user '%s' not found", username)
}

// Helper functions

func generateRandomSalt() ([]byte, error) {
	salt := make([]byte, 16) // 16 bytes = 128 bits
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func mechanismFromString(mechanism string) (sarama.ScramMechanismType, error) {
	switch strings.ToUpper(mechanism) {
	case "SCRAM-SHA-256":
		return sarama.SCRAM_MECHANISM_SHA_256, nil
	case "SCRAM-SHA-512":
		return sarama.SCRAM_MECHANISM_SHA_512, nil
	default:
		return sarama.SCRAM_MECHANISM_UNKNOWN, errors.Errorf("unsupported SCRAM mechanism: %s (supported: SCRAM-SHA-256, SCRAM-SHA-512)", mechanism)
	}
}

func mechanismToString(mechanism sarama.ScramMechanismType) string {
	switch mechanism {
	case sarama.SCRAM_MECHANISM_SHA_256:
		return "SCRAM-SHA-256"
	case sarama.SCRAM_MECHANISM_SHA_512:
		return "SCRAM-SHA-512"
	default:
		return "UNKNOWN"
	}
}
