package local

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	api_api "github.com/wunderkraut/radi-api/api"
	api_builder "github.com/wunderkraut/radi-api/builder"
	api_config "github.com/wunderkraut/radi-api/operation/config"

	handler_configwrapper "github.com/wunderkraut/radi-handlers/configwrapper"
	handler_local "github.com/wunderkraut/radi-handlers/local"
	handler_null "github.com/wunderkraut/radi-handlers/null"

	handler_libcompose_local "github.com/wunderkraut/radi-handler-libcompose/local"
	handler_rancher "github.com/wunderkraut/radi-handler-rancher"
	handler_upcloud "github.com/wunderkraut/radi-handler-upcloud"
)

/**
 * Make a local API
 *
 * First we check to see if we have a local project.  If not then
 * we return a minimal "build a project" api, which will give operations
 * but little real functionality outside of initializing a project
 *
 * If we have a project then we follow the following steps:
 *   1. build a bootstrap Project, which is used to load configuraion
 *   2. Use the configuration from the boostrap to build a real
 *      Project for the actual project config.
 *
 * This sequence is necessary to bypass the chicken-egg dilemna where
 * we need an Project to get project settings, but we need project settings
 * to decide what Project to build.  The Boostrap Project does have some weight,
 * but not much to worry about (perhaps a couple of files are opened 2x)
 *
 */
func MakeLocalAPI(settings handler_local.LocalAPISettings) (api_api.API, error) {

	if settings.ProjectDoesntExist {

		localProject, _ := MakeLocal_NoProject(settings)
		return localProject.API(), errors.New("No project found.")

	} else {

		// build a Project with at least the config operations, which we will need for a config wrapper
		log.Debug("Local:Project:: Building bootsrap Project")
		bootstrapProject := api_builder.New_StandardProject()
		bootstrapProject.AddBuilder(handler_local.New_LocalBuilder(settings))
		bootstrapProject.ActivateBuilder("local", *api_builder.New_Implementations([]string{"config"}), nil)

		// Now use the config operations to determine what builds are needed
		bootstrapOps := bootstrapProject.Operations()
		bootstrapConfigWrapper := api_config.New_SimpleConfigWrapper(bootstrapOps)

		localProject, err := MakeLocal_SecureProject(settings)

		ActivateConfigBuilders(localProject, settings, bootstrapConfigWrapper)

		validateResult := localProject.Validate()
		<-validateResult.Finished()
		if !validateResult.Success() {
			errs := validateResult.Errors()
			if len(errs) > 0 {
				err = errs[len(errs)-1]
			} else {
				err = errors.New("Secure local builder could not build a valid project")
			}
		}

		return localProject.API(), err
	}

}

// Add and activate Project builders from project config
func ActivateConfigBuilders(localProject api_builder.Project, localSettings handler_local.LocalAPISettings, configWrapper api_config.ConfigWrapper) {

	// Use the builderConfigWrapper to list build components
	log.Debug("CLI:LocalProject: Building SecureProject: Building builder configwrapper")
	projectConfigWrapper := handler_configwrapper.New_ProjectComponentsConfigWrapperYaml(configWrapper).ProjectConfigWrapper()
	builderList := projectConfigWrapper.List()
	log.WithFields(log.Fields{"list": builderList}).Debug("CLI:LocalProject: Building SecureProject: Built builder configwrapper")

	built := map[string]bool{}

	for _, key := range builderList {
		projectSetting, _ := projectConfigWrapper.Get(key)
		var buildErr error

		if _, checked := built[projectSetting.Type]; !checked {
			log.WithFields(log.Fields{"builder": projectSetting.Type}).Debug("CLI:LocalProject: Building SecureProject : AddingBuilder")
			built[projectSetting.Type] = true
			switch projectSetting.Type {
			case "null":
				log.Debug("CLI:LocalProject: Building Null builder")
				localProject.AddBuilder(handler_null.New_NullBuilder())
			case "local":
				log.Debug("CLI:LocalProject: Building Local builder")
				localProject.AddBuilder(handler_local.New_LocalBuilder(localSettings))
			case "upcloud":
				log.Debug("CLI:LocalProject: Building UpCloud builder")
				localProject.AddBuilder(api_builder.Builder(&handler_upcloud.UpcloudBuilder{}))
			case "rancher":
				log.Debug("CLI:LocalProject: Building Rancher builder")
				localProject.AddBuilder(api_builder.Builder(&handler_rancher.RancherBuilder{}))
			case "libcompose_local":
				log.Debug("CLI:LocalProject: Building Local builder")
				localProject.AddBuilder(handler_libcompose_local.New_LocalBuilder(localSettings))
			default:
				buildErr = errors.New("Unrecognized builder " + projectSetting.Type)
				log.WithError(buildErr).Error("Could not build " + projectSetting.Type)
				built[projectSetting.Type] = false
			}
		}

		if success, checked := built[projectSetting.Type]; success && checked {
			log.WithFields(log.Fields{"type": projectSetting.Type, "implementations": projectSetting.Implementations.Order(), "key": key}).Debug("CLI:LocalProject: Activate builder from settings")
			localProject.ActivateBuilder(projectSetting.Type, projectSetting.Implementations, projectSetting.SettingsProvider)
		} else {
			log.WithFields(log.Fields{"builder": projectSetting.Type}).Error("CLI:LocalProject: Unknown builder referenced in local project")
		}
	}

}
