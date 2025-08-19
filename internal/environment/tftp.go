package environment

import (
	"time"

	flag "github.com/namsral/flag"
)

// TFTPConfig carries runtime settings for the embedded TFTP server.
// It mirrors the CLI/config flags so the main package can decide whether to start it.
type TFTPConfig struct {
	Enabled  bool
	Addr     string
	Root     string
	Readonly bool
	Timeout  time.Duration
}

var (
	// These vars are bound to flags so the config file (key=value) and CLI both work.
	// Debian/Ubuntu packages document the "flags mirrored by config file" behavior,
	// so adding new flags here automatically enables config-file usage as well. See manpage.  ⟶ shoelaces(8)
	//   e.g. tftp-enabled=true
	//        tftp-addr=:69
	//        tftp-root=/var/lib/shoelaces/tftp
	//        tftp-readonly=true
	//        tftp-timeout=5s
	flagTftpEnabled  = flag.Bool("tftp-enabled", false, "Enable embedded TFTP server (serves iPXE loaders)")
	flagTftpAddr     = flag.String("tftp-addr", ":69", "TFTP listen address (UDP), e.g. 0.0.0.0:69")
	flagTftpRoot     = flag.String("tftp-root", "./tftp", "Directory to serve via TFTP (place undionly.kpxe/ipxe.efi here)")
	flagTftpReadonly = flag.Bool("tftp-readonly", true, "Disable TFTP uploads (recommended)")
	flagTftpTimeout  = flag.Duration("tftp-timeout", 5*time.Second, "Per-request TFTP timeout")
)

// tftpFromFlags returns a TFTPConfig from the currently parsed flags.
// Call this after your environment has parsed flags/config.
func tftpFromFlags() *TFTPConfig {
	return &TFTPConfig{
		Enabled:  *flagTftpEnabled,
		Addr:     *flagTftpAddr,
		Root:     *flagTftpRoot,
		Readonly: *flagTftpReadonly,
		Timeout:  *flagTftpTimeout,
	}
}
