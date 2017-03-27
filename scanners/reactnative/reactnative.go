package reactnative

import (
	"fmt"

	"path/filepath"

	"github.com/bitrise-core/bitrise-init/models"
	"github.com/bitrise-core/bitrise-init/scanners/android"
	"github.com/bitrise-core/bitrise-init/scanners/ios"
	"github.com/bitrise-core/bitrise-init/scanners/macos"
	"github.com/bitrise-core/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/log"
)

// ScannerName ...
const ScannerName = "reactnative"

const defaultConfigName = "default-reactnative-config"

const (
	projectPathKey    = "project_path"
	projectPathTitle  = "Project path"
	projectPathEnvKey = "BITRISE_PROJECT_PATH"

	schemeKey    = "scheme"
	schemeTitle  = "Scheme name"
	schemeEnvKey = "BITRISE_SCHEME"
)

// // ConfigDescriptor ...
// type configDescriptor struct {
// 	CanBuildAndroid  bool
// 	CanBuildiOS      bool
// }

// // func (descriptor ConfigDescriptor) String() string {
// // 	name := "reactnative-"
// // 	return name + "config"
// // }

// func (descriptor *configDescriptor) validate(scanner *Scanner) *configDescriptor {
// 	descriptor.CanBuildAndroid = (scanner.androidProjectDir != "" && scanner.androidProjectFile != "")
// 	descriptor.CanBuildiOS = (scanner.iOSProjectDir != "" && scanner.iOSProjectFile != "")
// 	descriptor.CanBundleAndroid = (scanner.androidProjectFile != "")
// 	descriptor.CanBundleiOS = (scanner.iOSProjectFile != "")
// 	descriptor.CanRunNpmTask = (scanner.packageJSONFile != "")
// 	return descriptor
// }

// Scanner ...
type Scanner struct {
	reactNativeProjectRootDir string
	searchDir                 string
	androidProjectFile        string
	iOSProjectFile            string
	androidProjectDir         string
	iOSProjectDir             string
	packageJSONFile           string
	iosScanner                *ios.Scanner
	androidScanner            *android.Scanner
}

// NewScanner ...
func NewScanner() *Scanner {
	return &Scanner{iosScanner: ios.NewScanner(), androidScanner: android.NewScanner()}
}

// Name ...
func (scanner Scanner) Name() string {
	return ScannerName
}

// DetectPlatform ...
func (scanner *Scanner) DetectPlatform(searchDir string) (bool, error) {
	scanner.searchDir = searchDir

	log.Infoft("Searching for React Native project files")

	//get all files in searchDir
	fileList, err := utility.ListPathInDirSortedByComponents(searchDir)
	if err != nil {
		return false, fmt.Errorf("failed to search for files in (%s), error: %s", searchDir, err)
	}

	reactNativeProjectFiles := []string{}

	//
	// Android
	// get android.index.js file
	androidProjectFiles, err := utility.FilterPaths(fileList,
		utility.AllowReactAndroidProjectBaseFilter,
		utility.ForbidReactTestsDir,
		utility.ForbidReactNodeModulesDir)
	if err != nil {
		return false, err
	}

	if len(androidProjectFiles) > 0 {
		// found android.index.js file, check if native project dir exists
		log.Printft("React Native android project file found")
		log.Printft("- %s", androidProjectFiles[0])

		if androidProjDir := utility.GetReactNativeAndroidProjectDirInDirectoryOf(androidProjectFiles[0]); androidProjDir != "" {
			log.Printft("Android project dir found")
			log.Printft("- %s", androidProjDir)
			log.Printft("")
			log.Infoft(">Run scanner: %s", android.ScannerName)

			if detected, err := scanner.androidScanner.DetectPlatform(androidProjDir); err != nil {
				return false, err
			} else if detected {
				scanner.androidProjectDir = androidProjDir
				reactNativeProjectFiles = append(reactNativeProjectFiles, androidProjectFiles[0])
			} else {
				log.Warnft("Android gradle file not found")
			}
		} else {
			log.Warnft("Android project dir not found")
		}

		scanner.androidProjectFile = androidProjectFiles[0]
	}

	//
	// iOS
	// check ios project JS
	iosProjectFiles, err := utility.FilterPaths(fileList,
		utility.AllowReactiOSProjectBaseFilter,
		utility.ForbidReactTestsDir,
		utility.ForbidReactNodeModulesDir)
	if err != nil {
		return false, err
	}
	if len(iosProjectFiles) > 0 {
		log.Printft("")
		log.Printft("React Native iOS project file found")
		log.Printft("- %s", iosProjectFiles[0])

		if iOSProjDir := utility.GetReactNativeiOSProjectDirInDirectoryOf(iosProjectFiles[0]); iOSProjDir != "" {
			log.Printft("iOS project dir found")
			log.Printft("- %s", iOSProjDir)
			log.Printft("")
			log.Infoft(">Run scanner: %s", ios.ScannerName)

			if detected, err := scanner.iosScanner.DetectPlatform(iOSProjDir); err != nil {
				return false, err
			} else if detected {
				scanner.iOSProjectDir = iOSProjDir
				reactNativeProjectFiles = append(reactNativeProjectFiles, iosProjectFiles[0])
			}
		}

		scanner.iOSProjectFile = iosProjectFiles[0]
	}

	packagesJSONFiles, err := utility.FilterPaths(fileList, utility.AllowReactNpmPackageBaseFilter)
	if err != nil {
		return false, err
	}
	if len(packagesJSONFiles) > 0 {
		scanner.packageJSONFile = packagesJSONFiles[0]
	}

	log.Printft("")
	log.Printft("%d React Native project files found", len(reactNativeProjectFiles))

	for _, reactNativeProjectFile := range reactNativeProjectFiles {
		log.Printft("- %s", reactNativeProjectFile)
	}

	if len(reactNativeProjectFiles) == 0 {
		log.Printft("Platform not detected")
		return false, nil
	}

	// get root project dir, and ensure projects are in the same dir
	if len(reactNativeProjectFiles) == 2 {
		projectRootDir := filepath.Dir(reactNativeProjectFiles[0])
		for _, reactNativeProjectFile := range reactNativeProjectFiles {
			projectFileDir := filepath.Dir(reactNativeProjectFile)
			if projectFileDir != projectRootDir {
				log.Errorft("React Native projects has different root directory")
				return false, nil
			}
		}
	}

	log.Doneft("Platform detected")
	return true, nil
}

// Options ...
func (scanner *Scanner) Options() (models.OptionModel, models.Warnings, error) {
	warnings := models.Warnings{}
	iosOptions, iosWarnings, err := scanner.iosScanner.Options()
	if err != nil {
		return models.OptionModel{}, models.Warnings{}, err
	}
	androidOptions, androidWarnings, err := scanner.androidScanner.Options()
	if err != nil {
		return models.OptionModel{}, models.Warnings{}, err
	}

	both := models.OptionModel{}

	projectPathOption := models.NewOptionModel("Platform", "")
	projectPathOption.ValueMap["iOS"] = iosOptions
	projectPathOption.ValueMap["Android"] = androidOptions
	projectPathOption.ValueMap["Both"] = both

	return projectPathOption, append(androidWarnings, iosWarnings...), nil

	optionID := ScannerName

	reactNativeTaskOption := models.NewOptionModel("React Native Task", "")

	buildConfig := models.NewEmptyOptionModel()

	isAndroidBuildAvailable := (scanner.androidProjectDir != "" && scanner.androidProjectFile != "")
	isIOSBuildAvailable := (scanner.iOSProjectDir != "" && scanner.iOSProjectFile != "")

	// add builds
	if isAndroidBuildAvailable || isIOSBuildAvailable {
		optionID += "-build"

		reactNativeBuildPlatformOption := models.NewOptionModel("Build Platform", "")

		if isIOSBuildAvailable {
			buildConfig.Config = optionID + "-ios"
			reactNativeBuildPlatformOption.ValueMap["iOS"] = buildConfig
		}
		if isAndroidBuildAvailable {
			buildConfig.Config = optionID + "-android"
			reactNativeBuildPlatformOption.ValueMap["Android"] = buildConfig
		}
		if isAndroidBuildAvailable && isIOSBuildAvailable {
			buildConfig.Config = optionID + "-ios-android"
			reactNativeBuildPlatformOption.ValueMap["iOS + Android"] = buildConfig
		}

		reactNativeTaskOption.ValueMap["Build"] = reactNativeBuildPlatformOption
	}

	optionID = ScannerName

	//add bundles
	if isAndroidBuildAvailable || isIOSBuildAvailable {
		optionID += "-bundle"

		reactNativeBundlePlatformOption := models.NewOptionModel("Bundle Platform", "")

		if isIOSBuildAvailable {
			buildConfig.Config = optionID + "-ios"
			reactNativeBundlePlatformOption.ValueMap["iOS"] = buildConfig
		}
		if isAndroidBuildAvailable {
			buildConfig.Config = optionID + "-android"
			reactNativeBundlePlatformOption.ValueMap["Android"] = buildConfig
		}
		if isAndroidBuildAvailable && isIOSBuildAvailable {
			buildConfig.Config = optionID + "-ios-android"
			reactNativeBundlePlatformOption.ValueMap["iOS + Android"] = buildConfig
		}

		reactNativeTaskOption.ValueMap["Bundle"] = reactNativeBundlePlatformOption
	}

	optionID = ScannerName

	//add tests
	if scanner.packageJSONFile != "" {
		optionID += "-test"
		reactNativeTestOption := models.NewOptionModel("Test", "")
		reactNativeTestOption.Config = optionID
		reactNativeTaskOption.ValueMap["Test"] = reactNativeTestOption
	}

	return reactNativeTaskOption, warnings, nil
}

// DefaultOptions ...
func (scanner *Scanner) DefaultOptions() models.OptionModel {
	return models.OptionModel{}
}

// Configs ...
func (scanner *Scanner) Configs() (models.BitriseConfigMap, error) {
	return models.BitriseConfigMap{}, nil
}

// DefaultConfigs ...
func (scanner *Scanner) DefaultConfigs() (models.BitriseConfigMap, error) {
	return models.BitriseConfigMap{}, nil
}

// IgnoreScanners ...
func (scanner *Scanner) IgnoreScanners() []string {
	isAndroidBuildAvailable := (scanner.androidProjectDir != "" && scanner.androidProjectFile != "")
	isIOSBuildAvailable := (scanner.iOSProjectDir != "" && scanner.iOSProjectFile != "")

	ignoreScanners := []string{}

	if isAndroidBuildAvailable {
		ignoreScanners = append(ignoreScanners, android.ScannerName)
	}

	if isIOSBuildAvailable {
		ignoreScanners = append(ignoreScanners, ios.ScannerName)
		ignoreScanners = append(ignoreScanners, macos.ScannerName)
	}

	return ignoreScanners
}
