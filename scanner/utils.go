package scanner

import (
	"errors"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/bitrise-core/bitrise-init/models"
	bitriseModels "github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/goinp/goinp"
)

func askForOptionValue(option models.OptionModel) (string, string, error) {
	optionValues := option.GetValues()

	selectedValue := ""
	if len(optionValues) == 1 {
		if optionValues[0] == "_" {
			// provide option value
			question := fmt.Sprintf("Provide: %s", option.Title)
			answer, err := goinp.AskForString(question)
			if err != nil {
				return "", "", err
			}

			selectedValue = answer
		} else {
			// auto select the only one value
			selectedValue = optionValues[0]
		}
	} else {
		// select from values
		question := fmt.Sprintf("Select: %s", option.Title)
		answer, err := goinp.SelectFromStrings(question, optionValues)
		if err != nil {
			return "", "", err
		}

		selectedValue = answer
	}

	return option.EnvKey, selectedValue, nil
}

// AskForOptions ...
func AskForOptions(options models.OptionModel) (string, []envmanModels.EnvironmentItemModel, error) {
	configPth := ""
	appEnvs := []envmanModels.EnvironmentItemModel{}

	var walkDepth func(option models.OptionModel) error

	walkDepth = func(option models.OptionModel) error {
		optionEnvKey, selectedValue, err := askForOptionValue(option)
		if err != nil {
			return fmt.Errorf("Failed to ask for vale, error: %s", err)
		}

		if optionEnvKey == "" {
			// last option selected, config got
			configPth = selectedValue
			return nil
		} else if optionEnvKey != "_" {
			// env's value selected
			appEnvs = append(appEnvs, envmanModels.EnvironmentItemModel{
				optionEnvKey: selectedValue,
			})
		}

		var nestedOptions *models.OptionModel
		if len(option.ChildOptionMap) == 1 {
			// auto select the next option
			for _, childOption := range option.ChildOptionMap {
				nestedOptions = childOption
				break
			}
		} else {
			// go to the next option, based on the selected value
			childOptions, found := option.ChildOptionMap[selectedValue]
			if !found {
				return nil
			}
			nestedOptions = childOptions
		}

		return walkDepth(*nestedOptions)
	}

	if err := walkDepth(options); err != nil {
		return "", []envmanModels.EnvironmentItemModel{}, err
	}

	if configPth == "" {
		return "", nil, errors.New("no config selected")
	}

	return configPth, appEnvs, nil
}

// AskForConfig ...
func AskForConfig(scanResult models.ScanResultModel) (bitriseModels.BitriseDataModel, error) {

	//
	// Select platform
	platforms := []string{}
	for platform := range scanResult.PlatformOptionMap {
		platforms = append(platforms, platform)
	}

	platform := ""
	if len(platforms) == 0 {
		return bitriseModels.BitriseDataModel{}, errors.New("no platform detected")
	} else if len(platforms) == 1 {
		platform = platforms[0]
	} else {
		var err error
		platform, err = goinp.SelectFromStrings("Select platform", platforms)
		if err != nil {
			return bitriseModels.BitriseDataModel{}, err
		}
	}
	// ---

	//
	// Select config
	options, ok := scanResult.PlatformOptionMap[platform]
	if !ok {
		return bitriseModels.BitriseDataModel{}, fmt.Errorf("invalid platform selected: %s", platform)
	}

	configPth, appEnvs, err := AskForOptions(options)
	if err != nil {
		return bitriseModels.BitriseDataModel{}, err
	}
	// --

	//
	// Build config
	configMap := scanResult.PlatformConfigMapMap[platform]
	configStr := configMap[configPth]

	var config bitriseModels.BitriseDataModel
	if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
		return bitriseModels.BitriseDataModel{}, fmt.Errorf("failed to unmarshal config, error: %s", err)
	}

	config.App.Environments = append(config.App.Environments, appEnvs...)
	// ---

	return config, nil
}
