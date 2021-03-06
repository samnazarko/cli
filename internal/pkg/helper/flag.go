package helper

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ukfast/cli/internal/pkg/clierrors"
	"github.com/ukfast/sdk-go/pkg/connection"
)

// InferTypeFromStringFlag will return a int, bool or string, based on value of flag
func InferTypeFromStringFlag(flag string) interface{} {
	intValue, err := strconv.Atoi(flag)
	if err == nil {
		return intValue
	}

	boolValue, err := strconv.ParseBool(flag)
	if err == nil {
		return boolValue
	}

	return flag
}

// GetFilteringInferOperator returns an APIRequestFiltering struct with the operater inferred from the
// input value. This will return an operator based on the following rules:
// If value contains a comma ',' - returns IN operator
// If value contains an asterisk '*' - returns LK operator
// Otherwise returns EQ operator
func GetFilteringInferOperator(property string, value string) connection.APIRequestFiltering {
	return connection.APIRequestFiltering{
		Property: property,
		Operator: inferOperatorFromValue(value),
		Value:    []string{value},
	}
}

func inferOperatorFromValue(value string) connection.APIRequestFilteringOperator {
	if strings.Contains(value, ",") {
		return connection.INOperator
	}
	if strings.Contains(value, "*") {
		return connection.LKOperator
	}

	return connection.EQOperator
}

// GetFilteringArrayFromStringArrayFlag retrieves an array of APIRequestFiltering structs for given
// filtering strings
func GetFilteringArrayFromStringArrayFlag(filters []string) ([]connection.APIRequestFiltering, error) {
	var filtering []connection.APIRequestFiltering
	for _, filter := range filters {
		f, err := GetFilteringFromStringFlag(filter)
		if err != nil {
			return filtering, clierrors.NewErrInvalidFlagValue("filter", filter, err)
		}

		filtering = append(filtering, f)
	}

	return filtering, nil
}

// GetFilteringFromStringFlag retrieves a APIRequestFiltering struct from given filtering
// string. This function expects a string in the following format (with optional :operator): propertyname:operator=value,
// Valid examples:
// name:eq=something
// name=something
func GetFilteringFromStringFlag(filter string) (connection.APIRequestFiltering, error) {
	filtering := connection.APIRequestFiltering{}

	if filter == "" {
		return filtering, nil
	}

	// Obtain KV parts from filtering flag string. Example: propertyname:eq=value
	// K at index 0 represents propertyname and optional :operator in format propertyname:operator
	// V at index 1 represents the value
	filteringKVParts := strings.Split(filter, "=")
	if len(filteringKVParts) != 2 || filteringKVParts[1] == "" {
		return filtering, errors.New("Missing value for filtering")
	}

	// Obtain PropertyOperator parts from K above. Example: propertyname:operator
	// Property at index 0 represents the property name
	// Operator at index 1 represents the operator
	filteringPropertyOperatorParts := strings.Split(filteringKVParts[0], ":")
	if filteringPropertyOperatorParts[0] == "" {
		return filtering, errors.New("Missing property for filtering")
	}

	var operator connection.APIRequestFilteringOperator
	if len(filteringPropertyOperatorParts) == 1 {
		operator = inferOperatorFromValue(filteringKVParts[1])
	} else {
		if len(filteringPropertyOperatorParts) != 2 || filteringPropertyOperatorParts[1] == "" {
			return filtering, errors.New("Missing operator for filtering")
		}

		// Parse the operator, returning parse error if any
		parsedOperator, err := connection.ParseOperator(filteringPropertyOperatorParts[1])
		if err != nil {
			return filtering, err
		}

		operator = parsedOperator
	}

	// Sanitize comma-separated value by trimming spaces following split
	var sanitizedValues []string
	values := strings.Split(filteringKVParts[1], ",")
	for _, value := range values {
		sanitizedValues = append(sanitizedValues, strings.TrimSpace(value))
	}

	filtering.Property = filteringPropertyOperatorParts[0]
	filtering.Operator = operator
	filtering.Value = sanitizedValues

	return filtering, nil
}

// GetSortingFromStringFlag return an APIRequestSorting struct from given sorting string flag
func GetSortingFromStringFlag(sort string) connection.APIRequestSorting {
	if sort == "" {
		return connection.APIRequestSorting{}
	}

	var descending bool

	sortingParts := strings.Split(sort, ":")
	if (len(sortingParts)) > 1 && strings.ToLower(sortingParts[1]) == "desc" {
		descending = true
	}

	return connection.APIRequestSorting{
		Property:   sortingParts[0],
		Descending: descending,
	}
}

func GetAPIRequestParametersFromFlags(cmd *cobra.Command) (connection.APIRequestParameters, error) {
	flagFilter, _ := cmd.Flags().GetStringArray("filter")
	filtering, err := GetFilteringArrayFromStringArrayFlag(flagFilter)
	if err != nil {
		return connection.APIRequestParameters{}, err
	}

	flagSort, err := cmd.Flags().GetString("sort")

	return connection.APIRequestParameters{
		Sorting:   GetSortingFromStringFlag(flagSort),
		Filtering: filtering,
		Pagination: connection.APIRequestPagination{
			PerPage: viper.GetInt("api_pagination_perpage"),
		},
	}, nil
}

func GetContentsFromFilePathFlag(cmd *cobra.Command, fs afero.Fs, filePathFlag string) (string, error) {
	filePath, _ := cmd.Flags().GetString(filePathFlag)
	file, err := fs.Open(filePath)
	if err != nil {
		return "", err
	}

	contentBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(contentBytes), nil
}
