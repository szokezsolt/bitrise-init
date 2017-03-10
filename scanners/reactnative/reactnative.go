package reactnative

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-core/bitrise-init/models"
	"github.com/bitrise-core/bitrise-init/steps"
	"github.com/bitrise-core/bitrise-init/utility"
	bitriseModels "github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/log"
)

const scannerName = "reactnative"

const defaultConfigName = "default-reactnative-config"

const (
	projectPathKey    = "project_path"
	projectPathTitle  = "Project path"
	projectPathEnvKey = "BITRISE_PROJECT_PATH"

	schemeKey    = "scheme"
	schemeTitle  = "Scheme name"
	schemeEnvKey = "BITRISE_SCHEME"
)

// ConfigDescriptor ...
type configDescriptor struct {
	CanBuildAndroid  bool
	CanBuildiOS      bool
	CanBundleAndroid bool
	CanBundleiOS     bool
	CanRunNpmTask    bool
}

// func (descriptor ConfigDescriptor) String() string {
// 	name := "reactnative-"
// 	return name + "config"
// }

func (descriptor *configDescriptor) validate(scanner *Scanner) *configDescriptor {
	descriptor.CanBuildAndroid = (scanner.androidProjectDir != "" && scanner.androidProjectFile != "")
	descriptor.CanBuildiOS = (scanner.iOSProjectDir != "" && scanner.iOSProjectFile != "")
	descriptor.CanBundleAndroid = (scanner.androidProjectFile != "")
	descriptor.CanBundleiOS = (scanner.iOSProjectFile != "")
	descriptor.CanRunNpmTask = (scanner.packageJSONFile != "")
	return descriptor
}

// Scanner ...
type Scanner struct {
	searchDir          string
	fileList           []string
	androidProjectFile string
	iOSProjectFile     string
	androidProjectDir  string
	iOSProjectDir      string
	packageJSONFile    string
}

// Name ...
func (scanner Scanner) Name() string {
	return scannerName
}

// Print ...
func (scanner Scanner) Print() {
	log.Printft("searchDir: %s", scanner.searchDir)
	log.Printft("androidProjectFile: %s", scanner.androidProjectFile)
	log.Printft("iOSProjectFile: %s", scanner.iOSProjectFile)
	log.Printft("androidProjectDir: %s", scanner.androidProjectDir)
	log.Printft("iOSProjectDir: %s", scanner.iOSProjectDir)
	log.Printft("packageJSONFile: %s", scanner.packageJSONFile)
}

// DetectPlatform ...
func (scanner *Scanner) DetectPlatform(searchDir string) (bool, error) {
	scanner.searchDir = searchDir

	fileList, err := utility.ListPathInDirSortedByComponents(searchDir, true)
	if err != nil {
		return false, fmt.Errorf("failed to search for files in (%s), error: %s", searchDir, err)
	}
	scanner.fileList = fileList

	// Search for xcodeproj
	log.Infoft("Searching for React Native project files")

	reactNativeProjectFiles := []string{}

	// check android project JS and native android dir
	androidProjectFiles, err := utility.FilterPaths(fileList,
		utility.AllowReactAndroidProjectBaseFilter,
		utility.ForbidReactTestsDir,
		utility.ForbidReactNodeModulesDir)
	if err != nil {
		return false, err
	}
	if len(androidProjectFiles) > 0 {
		if androidProjDir := utility.GetReactNativeAndroidProjectDirInDirectoryOf(androidProjectFiles[0]); androidProjDir != "" {
			scanner.androidProjectDir = androidProjDir
		}
		scanner.androidProjectFile = androidProjectFiles[0]
	}
	reactNativeProjectFiles = append(reactNativeProjectFiles, androidProjectFiles...)

	// check ios project JS and native android dir
	iosProjectFiles, err := utility.FilterPaths(fileList,
		utility.AllowReactiOSProjectBaseFilter,
		utility.ForbidReactTestsDir,
		utility.ForbidReactNodeModulesDir)
	if err != nil {
		return false, err
	}
	if len(iosProjectFiles) > 0 {
		if iOSProjDir := utility.GetReactNativeiOSProjectDirInDirectoryOf(iosProjectFiles[0]); iOSProjDir != "" {
			scanner.iOSProjectDir = iOSProjDir
		}
		scanner.iOSProjectFile = iosProjectFiles[0]
	}
	reactNativeProjectFiles = append(reactNativeProjectFiles, iosProjectFiles...)

	packagesJSONFiles, err := utility.FilterPaths(fileList, utility.AllowReactNpmPackageBaseFilter)
	if err != nil {
		return false, err
	}
	if len(packagesJSONFiles) > 0 {
		scanner.packageJSONFile = packagesJSONFiles[0]
	}

	log.Printft("%d React Native project files found", len(reactNativeProjectFiles))

	for _, reactNativeProjectFile := range reactNativeProjectFiles {
		log.Printft("- %s", reactNativeProjectFile)
	}

	if len(reactNativeProjectFiles) == 0 {
		log.Printft("Platform not detected")
		return false, nil
	}
	scanner.Print()
	log.Doneft("Platform detected")
	return true, nil
}

// Options ...
func (scanner *Scanner) Options() (models.OptionModel, models.Warnings, error) {
	warnings := models.Warnings{}

	descriptor := configDescriptor{}
	descriptor.validate(scanner)

	optionID := scannerName

	reactNativeTaskOption := models.NewOptionModel("React Native Task", "")

	buildConfig := models.NewEmptyOptionModel()

	// add builds
	if descriptor.CanBuildAndroid || descriptor.CanBuildiOS {
		optionID += "-build"

		reactNativeBuildPlatformOption := models.NewOptionModel("Build Platform", "")

		if descriptor.CanBuildiOS {
			buildConfig.Config = optionID + "-ios"
			reactNativeBuildPlatformOption.ValueMap["iOS"] = buildConfig
		}
		if descriptor.CanBuildAndroid {
			buildConfig.Config = optionID + "-android"
			reactNativeBuildPlatformOption.ValueMap["Android"] = buildConfig
		}
		if descriptor.CanBuildAndroid && descriptor.CanBuildiOS {
			buildConfig.Config = optionID + "-ios-android"
			reactNativeBuildPlatformOption.ValueMap["iOS + Android"] = buildConfig
		}

		reactNativeTaskOption.ValueMap["Build"] = reactNativeBuildPlatformOption
	}

	optionID = scannerName

	//add bundles
	if descriptor.CanBundleAndroid || descriptor.CanBundleiOS {
		optionID += "-bundle"

		reactNativeBundlePlatformOption := models.NewOptionModel("Bundle Platform", "")

		if descriptor.CanBundleiOS {
			buildConfig.Config = optionID + "-ios"
			reactNativeBundlePlatformOption.ValueMap["iOS"] = buildConfig
		}
		if descriptor.CanBundleAndroid {
			buildConfig.Config = optionID + "-android"
			reactNativeBundlePlatformOption.ValueMap["Android"] = buildConfig
		}
		if descriptor.CanBundleAndroid && descriptor.CanBundleiOS {
			buildConfig.Config = optionID + "-ios-android"
			reactNativeBundlePlatformOption.ValueMap["iOS + Android"] = buildConfig
		}

		reactNativeTaskOption.ValueMap["Bundle"] = reactNativeBundlePlatformOption
	}

	optionID = scannerName

	//add tests
	if descriptor.CanRunNpmTask {
		optionID += "-test"
		reactNativeTestOption := models.NewOptionModel("Test", "")
		reactNativeTestOption.Config = optionID
		reactNativeTaskOption.ValueMap["Test"] = reactNativeTestOption
	}

	return reactNativeTaskOption, warnings, nil
}

// DefaultOptions ...
func (scanner *Scanner) DefaultOptions() models.OptionModel {
	configOption := models.NewEmptyOptionModel()
	configOption.Config = defaultConfigName

	projectPathOption := models.NewOptionModel(projectPathTitle, projectPathEnvKey)
	schemeOption := models.NewOptionModel(schemeTitle, schemeEnvKey)

	schemeOption.ValueMap["_"] = configOption
	projectPathOption.ValueMap["_"] = schemeOption

	return projectPathOption
}

// Configs ...
func (scanner *Scanner) Configs() (models.BitriseConfigMap, error) {

	bitriseDataMap := models.BitriseConfigMap{}

	return bitriseDataMap, nil
}

// DefaultConfigs ...
func (scanner *Scanner) DefaultConfigs() (models.BitriseConfigMap, error) {
	//
	// Prepare steps
	prepareSteps := []bitriseModels.StepListItemModel{}

	// ActivateSSHKey
	prepareSteps = append(prepareSteps, steps.ActivateSSHKeyStepListItem())

	// GitClone
	prepareSteps = append(prepareSteps, steps.GitCloneStepListItem())

	// Script
	prepareSteps = append(prepareSteps, steps.ScriptSteplistItem(steps.ScriptDefaultTitle))

	// CertificateAndProfileInstaller
	prepareSteps = append(prepareSteps, steps.CertificateAndProfileInstallerStepListItem())

	// CocoapodsInstall
	prepareSteps = append(prepareSteps, steps.CocoapodsInstallStepListItem())

	// RecreateUserSchemes
	prepareSteps = append(prepareSteps, steps.RecreateUserSchemesStepListItem([]envmanModels.EnvironmentItemModel{
		envmanModels.EnvironmentItemModel{projectPathKey: "$" + projectPathEnvKey},
	}))
	// ----------

	//
	// CI steps
	ciSteps := append([]bitriseModels.StepListItemModel{}, prepareSteps...)

	// XcodeTest
	ciSteps = append(ciSteps, steps.XcodeTestStepListItem([]envmanModels.EnvironmentItemModel{
		envmanModels.EnvironmentItemModel{projectPathKey: "$" + projectPathEnvKey},
		envmanModels.EnvironmentItemModel{schemeKey: "$" + schemeEnvKey},
	}))

	// DeployToBitriseIo
	ciSteps = append(ciSteps, steps.DeployToBitriseIoStepListItem())
	// ----------

	//
	// Deploy steps
	deploySteps := append([]bitriseModels.StepListItemModel{}, prepareSteps...)

	// XcodeTest
	deploySteps = append(deploySteps, steps.XcodeTestStepListItem([]envmanModels.EnvironmentItemModel{
		envmanModels.EnvironmentItemModel{projectPathKey: "$" + projectPathEnvKey},
		envmanModels.EnvironmentItemModel{schemeKey: "$" + schemeEnvKey},
	}))

	// XcodeArchive
	deploySteps = append(deploySteps, steps.XcodeArchiveStepListItem([]envmanModels.EnvironmentItemModel{
		envmanModels.EnvironmentItemModel{projectPathKey: "$" + projectPathEnvKey},
		envmanModels.EnvironmentItemModel{schemeKey: "$" + schemeEnvKey},
	}))

	// DeployToBitriseIo
	deploySteps = append(deploySteps, steps.DeployToBitriseIoStepListItem())
	// ----------

	config := models.BitriseDataWithCIAndCDWorkflow([]envmanModels.EnvironmentItemModel{}, ciSteps, deploySteps)
	data, err := yaml.Marshal(config)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	configName := defaultConfigName
	bitriseDataMap := models.BitriseConfigMap{}
	bitriseDataMap[configName] = string(data)

	return bitriseDataMap, nil
}
