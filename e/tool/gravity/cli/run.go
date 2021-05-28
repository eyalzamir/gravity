// Copyright 2021 Gravitational Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"os"

	"github.com/gravitational/gravity/e/lib/environment"
	"github.com/gravitational/gravity/lib/state"
	cliutils "github.com/gravitational/gravity/lib/utils/cli"
	"github.com/gravitational/gravity/tool/gravity/cli"

	"github.com/gravitational/configure/cstrings"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField(trace.Component, "cli")

func Run(g *Application) error {
	log.WithField("args", os.Args).Debug("Executing command.")
	err := cli.ConfigureEnvironment()
	if err != nil {
		return trace.Wrap(err)
	}

	args, extraArgs := cstrings.SplitAt(os.Args[1:], "--")
	cmd, err := g.Parse(args)
	if err != nil {
		return trace.Wrap(err)
	}

	if *g.UID != -1 || *g.GID != -1 {
		return cli.SwitchPrivileges(*g.UID, *g.GID)
	}

	err = cli.InitAndCheck(g.Application, cmd)
	if err != nil {
		return trace.Wrap(err)
	}

	execer := cli.CmdExecer{
		Exe:       getExec(g, cmd, extraArgs),
		Parser:    cliutils.ArgsParserFunc(parseArgs),
		Args:      args,
		ExtraArgs: extraArgs,
	}
	return execer.Execute()
}

func getExec(g *Application, cmd string, extraArgs []string) cli.Executable {
	return func() error {
		return execute(g, cmd, extraArgs)
	}
}

func execute(g *Application, cmd string, extraArgs []string) (err error) {
	switch cmd {
	case g.SiteStartCmd.FullCommand():
		return startProcess(
			*g.SiteStartCmd.ConfigPath,
			*g.SiteStartCmd.InitPath)
	}

	// the following enterprise commands require local environment
	var localEnv *environment.Local
	switch cmd {
	case g.InstallCmd.FullCommand():
		if *g.StateDir != "" {
			if err := state.SetStateDir(*g.StateDir); err != nil {
				return trace.Wrap(err)
			}
		}
		ossLocalEnv, err := g.NewInstallEnv()
		if err != nil {
			return trace.Wrap(err)
		}
		localEnv = &environment.Local{LocalEnvironment: ossLocalEnv}
		defer localEnv.Close()
	case g.WizardCmd.FullCommand(),
		g.StatusCmd.FullCommand(),
		g.UpdateDownloadCmd.FullCommand(),
		g.OpsGenerateCmd.FullCommand(),
		g.TunnelEnableCmd.FullCommand(),
		g.TunnelDisableCmd.FullCommand(),
		g.TunnelStatusCmd.FullCommand(),
		g.ResourceCreateCmd.FullCommand(),
		g.ResourceRemoveCmd.FullCommand(),
		g.ResourceGetCmd.FullCommand(),
		g.LicenseInstallCmd.FullCommand(),
		g.LicenseNewCmd.FullCommand(),
		g.LicenseShowCmd.FullCommand(),
		g.SiteInfoCmd.FullCommand():
		ossLocalEnv, err := g.NewLocalEnv()
		if err != nil {
			return trace.Wrap(err)
		}
		localEnv = &environment.Local{LocalEnvironment: ossLocalEnv}
		defer localEnv.Close()
	}
	switch cmd {
	case g.InstallCmd.FullCommand():
		config, err := NewInstallConfig(localEnv.LocalEnvironment, g)
		if err != nil {
			return trace.Wrap(err)
		}
		return startInstall(localEnv, *config)
	case g.WizardCmd.FullCommand():
		config, err := newWizardConfig(localEnv, g)
		if err != nil {
			return trace.Wrap(err)
		}
		return startInstall(localEnv, *config)
	case g.StatusCmd.FullCommand():
		// only --tunnel flag is specific to the enterprise
		if *g.StatusCmd.Tunnel {
			return remoteAccessStatus(localEnv)
		}
	case g.UpdateDownloadCmd.FullCommand():
		return updateDownload(localEnv, *g.UpdateDownloadCmd.Every)
	case g.OpsGenerateCmd.FullCommand():
		return generateInstaller(localEnv,
			*g.OpsGenerateCmd.Package,
			*g.OpsGenerateCmd.Dir,
			*g.OpsGenerateCmd.CACert,
			*g.OpsGenerateCmd.EncryptionKey,
			*g.OpsGenerateCmd.OpsCenterURL)
	case g.TunnelEnableCmd.FullCommand():
		return updateRemoteAccess(localEnv, true)
	case g.TunnelDisableCmd.FullCommand():
		return updateRemoteAccess(localEnv, false)
	case g.TunnelStatusCmd.FullCommand():
		return remoteAccessStatus(localEnv)
	case g.ResourceCreateCmd.FullCommand():
		return createResource(localEnv, g.Application,
			*g.ResourceCreateCmd.Filename,
			*g.ResourceCreateCmd.Upsert,
			*g.ResourceCreateCmd.User,
			*g.ResourceCreateCmd.Manual,
			*g.ResourceCreateCmd.Confirmed)
	case g.ResourceRemoveCmd.FullCommand():
		return removeResource(localEnv, g.Application,
			*g.ResourceRemoveCmd.Kind,
			*g.ResourceRemoveCmd.Name,
			*g.ResourceRemoveCmd.Force,
			*g.ResourceRemoveCmd.User,
			*g.ResourceRemoveCmd.Manual,
			*g.ResourceRemoveCmd.Confirmed)
	case g.ResourceGetCmd.FullCommand():
		return getResources(localEnv,
			*g.ResourceGetCmd.Kind,
			*g.ResourceGetCmd.Name,
			*g.ResourceGetCmd.WithSecrets,
			*g.ResourceGetCmd.Format,
			*g.ResourceGetCmd.User)
	case g.LicenseInstallCmd.FullCommand():
		return installLicense(localEnv,
			*g.LicenseInstallCmd.Path)
	case g.LicenseNewCmd.FullCommand():
		return newLicense(
			*g.LicenseNewCmd.MaxNodes,
			*g.LicenseNewCmd.ValidFor,
			*g.LicenseNewCmd.StopApp,
			*g.LicenseNewCmd.CAKey,
			*g.LicenseNewCmd.CACert,
			*g.LicenseNewCmd.EncryptionKey,
			*g.LicenseNewCmd.CustomerName,
			*g.LicenseNewCmd.CustomerEmail,
			*g.LicenseNewCmd.CustomerMetadata,
			*g.LicenseNewCmd.ProductName,
			*g.LicenseNewCmd.ProductVersion)
	case g.LicenseShowCmd.FullCommand():
		return showLicense(localEnv,
			*g.LicenseShowCmd.Output)
	case g.SiteInfoCmd.FullCommand():
		return printLocalClusterInfo(localEnv,
			*g.SiteInfoCmd.Format)
	}
	// no enterprise commands matched, execute open-source
	return cli.Execute(g.Application, cmd, extraArgs)
}
