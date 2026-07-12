// Deterministic blob example: bit-identical binary exports.
//
// Recension stores these in MinIO and compares by sha256 digest only.
// Same generator inputs → same bytes → pass. One flipped byte → diff.
//
// This is the right MinIO use case (firmware/export/protobuf snapshots),
// not "PDF that looks the same but was re-exported".
package main

import (
	"encoding/binary"
	"flag"
	"hash/crc32"
	"os"

	recension "github.com/ayitas/recension/sdk/go"
)

// exportArtifact builds a tiny deterministic binary container:
// magic(4) + name + payload + crc32.
// No timestamps, random IDs, or wall-clock — re-runs yield identical bytes.
func exportArtifact(name string, breakExport bool) []byte {
	payload := map[string][]byte{
		"invoice": {0x01, 0x02, 0x03, 0x04, 0x10, 0x20},
		"receipt": {0xaa, 0xbb, 0xcc, 0x11, 0x22, 0x33},
	}[name]
	if payload == nil {
		payload = []byte(name)
	}
	if breakExport {
		// Simulate an unintentional generator change (one byte).
		out := append([]byte(nil), payload...)
		out[0] ^= 0xff
		payload = out
	}

	buf := make([]byte, 0, 64)
	buf = append(buf, 'R', 'C', 'N', '1') // magic
	buf = append(buf, byte(len(name)))
	buf = append(buf, name...)
	buf = append(buf, byte(len(payload)))
	buf = append(buf, payload...)
	sum := crc32.ChecksumIEEE(buf)
	var crc [4]byte
	binary.BigEndian.PutUint32(crc[:], sum)
	return append(buf, crc[:]...)
}

func main() {
	breakExport := flag.Bool("break", false, "flip one payload byte to force a digest diff")
	// Do not flag.Parse() here — recension.Run registers its flags then parses once.

	recension.Workflow("exports", func(name string) {
		recension.Assume("name", name)
		recension.StartTimer("export_artifact")
		blob := recension.BlobFile{
			Data: exportArtifact(name, *breakExport),
			Mime: "application/octet-stream",
		}
		recension.StopTimer("export_artifact")
		recension.Check("export.bin", blob)
	}, recension.WithTestcases([]string{"invoice", "receipt"}))

	os.Exit(recension.Run())
}
