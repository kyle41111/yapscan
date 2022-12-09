package app

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/fkie-cad/yapscan/archiver"

	"github.com/fkie-cad/yapscan/pgp"

	"github.com/fkie-cad/yapscan/output"
	"github.com/targodan/go-errors"
	"github.com/urfave/cli/v2"
)

func receive(c *cli.Context) error {
	err := initAppAction(c)
	if err != nil {
		return err
	}

	if c.NArg() != 1 {
		return errors.Newf("expected exactly one argument <listen address>")
	}

	if c.Bool("verbose") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	wcBuilder := output.NewWriteCloserBuilder()
	if c.String("password") != "" && c.String("pgpkey") != "" {
		return fmt.Errorf("cannot encrypt with both pgp key and a password")
	}
	if c.String("password") != "" {
		wcBuilder.Append(output.PGPSymmetricEncryptionDecorator(c.String("password"), true))
	}
	if c.String("pgpkey") != "" {
		ring, err := pgp.ReadKeyRing(c.String("pgpkey"))
		if err != nil {
			return fmt.Errorf("could not read specified public pgp key, reason: %w", err)
		}
		wcBuilder.Append(output.PGPEncryptionDecorator(ring, true))
	}
	wcBuilder.Append(output.ZSTDCompressionDecorator())

	reportServer := archiver.NewArchiverServer(c.String("report-dir"), wcBuilder.SuggestedFileExtension(), wcBuilder)
	return reportServer.Start(c.Args().First())
}
