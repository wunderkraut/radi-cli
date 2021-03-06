package main

import (
	"errors"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"

	api_operation "github.com/wunderkraut/radi-api/operation"
	api_security "github.com/wunderkraut/radi-api/operation/security"
	api_property "github.com/wunderkraut/radi-api/property"
)

// Add operations from the API to the app
func AppApiOperations(app *cli.App, ops api_operation.Operations, internal bool) error {
	for _, id := range ops.Order() {
		op, _ := ops.Get(id)

		log.WithFields(log.Fields{"id": op.Id()}).Debug("Operation: " + op.Label())
		// we could also add "label": op.Label(), "description": op.Description(), "configurations": op.Properties()

		// usage := op.Usage()
		// log.WithFields(log.Fields{"id": id, "usage": usage, "is-external": api_operation.IsUsage_External(usage)}).Info("Operation usage investigation")

		if internal || api_operation.IsUsage_External(op.Usage()) {
			id := op.Id()
			category := id[0:strings.Index(id, ".")]
			alias := id[strings.LastIndex(id, ".")+1:]

			log.WithFields(log.Fields{"id": id, "category": category, "alias": alias}).Debug("Cli: Adding Operation")

			opWrapper := CliOperationWrapper{
				op:       op,
				internal: internal,
			}
			cliComm := cli.Command{
				Name:     op.Id(),
				Aliases:  []string{alias},
				Usage:    op.Description(),
				Action:   opWrapper.Exec,
				Category: category,
			}

			cliComm.Flags = CliMakeFlagsFromProperties(op.Properties(), internal)

			app.Commands = append(app.Commands, &cliComm)
		}
	}

	return nil
}

/**
 * Wrapper for operation Exec methods, from the urface CLI
 *
 * We use this wrapper because:
 *  1. the cli library has a different return
 *     expectation than what our operations return
 *  2. we need to do some minor transformation on CLI
 *     arguments, to make them fit our types.
 *  3. we want to do some work to decide what to output
 *     to the screen.
 */
type CliOperationWrapper struct {
	op       api_operation.Operation
	internal bool
}

// Execute the operation for the cli
func (opWrapper *CliOperationWrapper) Exec(cliContext *cli.Context) error {
	logger := log.WithFields(log.Fields{"id": opWrapper.op.Id()})
	logger.Debug("Running operation")

	props := opWrapper.op.Properties()

	CliAssignPropertiesFromFlags(cliContext, props, opWrapper.internal)

	result := opWrapper.op.Exec(props)
	<-result.Finished()

	success := result.Success() // bool
	errs := result.Errors()     // []error

	if !success && len(errs) == 0 {
		errs = []error{errors.New("RadiCLI: Unknown error occured")}
	}

	// Create some meaningful output, by logging some of the properties
	fields := map[string]interface{}{
		"success": success,
		"errors":  errs,
	}

	logger = logger.WithFields(log.Fields(fields))

	if success {
		for _, key := range props.Order() {
			prop, _ := props.Get(key)

			log.WithFields(log.Fields{"id": prop.Id(), "type": prop.Type(), "value": prop.Get()}).Debug("CLI:Operation: Properties collect")

			if opWrapper.internal || api_property.IsUsage_ExternalVisibleAfter(prop.Usage()) {
				switch prop.Type() {
				case "string":
					fields[key] = prop.Get().(string)
				case "[]string":
					fields[key] = prop.Get().([]string)
				case "[]byte":
					fields[key] = string(prop.Get().([]byte))
				case "int32":
					fields[key] = int(prop.Get().(int32))
				case "int64":
					fields[key] = prop.Get().(int64)
				case "bool":
					fields[key] = prop.Get().(bool)
				case "github.com/wunderkraut/radi-api/operation/security.SecurityUser":
					user := prop.Get().(api_security.SecurityUser)
					fields[key] = user.Id()
				}
			}
		}

		logger = logger.WithFields(log.Fields(fields))
		logger.Info("Operation completed.")
		return nil
	} else {
		for _, err := range errs {
			logger = logger.WithError(err)
		}

		logger.Error("Error occured running operation")
		if len(errs) > 0 {
			return errs[len(errs)-1]
		} else {
			return errors.New("Unknown error occured when trying to run an operation")
		}
	}
}
