package geoip

import (
	"archive/tar"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
)

type MaxmindGeoIP2 struct {
	sync.Mutex
	reader    *geoip2.Reader
	reloading bool
	lastTried     time.Time
}

var geoIP MaxmindGeoIP2

func loadGeoDB(Path string, geoUrl string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error in geoip: %s", r)
		}
	}()
	var g = &geoIP
	var dbPath string
	dbPath = filepath.Join(Path, "country.mmdb")
	stat, err := os.Stat(dbPath)
	reload := false
	if err != nil || stat.ModTime().Unix() < time.Now().Add(-24*time.Hour).Unix() {
		reload = true
	}
	exist := false
	if g.reader != nil {
		exist = true
	}
	if !exist {
		var reader *geoip2.Reader
		reader, err = geoip2.Open(dbPath)
		if err != nil {
			log.Printf("Can't open maxmind database: %s", err)
			reload = true
		} else {
			testIp := net.ParseIP("8.8.8.8")
			g.Lock()
			_, err = reader.Country(testIp)
			if err != nil {
				reload = true
			} else {
				g.reader = reader
			}
			g.Unlock()
			log.Printf("Geo DB %s initialized", filepath.Base(dbPath))
		}
	}
	if reload {
		g.Lock()
		if g.reloading {
			g.Unlock()
			return
		}
		g.reloading = true
		g.lastTried = time.Now()
		g.Unlock()
		defer func() {
			g.Lock()
			g.reloading = false
			g.Unlock()
		}()
		// Let's download GeoIP from the source
		url := geoUrl
		if url == "" {
			url = "https://totaltraffictrader.com/geo/country.tar.gz"
		}
		var archiveFile *os.File
		archiveFile, err = os.Create(dbPath + ".tar.gz")
		if err != nil {
			log.Println("can't create file", dbPath+".tar.gz")
			return
		}
		log.Printf("loading geoIP database from %s...", url)
		var resp *http.Response
		resp, err = http.Get(url)
		if err != nil {
			log.Println("error downloading geoip database", err)
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "application/gzip" &&
			resp.Header.Get("Content-Type") != "application/x-gzip" {
			log.Printf("Wrong content type from %s: %s", url, resp.Header.Get("Content-Type"))
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}

		_, err = io.Copy(archiveFile, resp.Body)
		if err != nil {
			log.Println("can't save geoip archive to", dbPath+".tar.gz")
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}

		_, err = archiveFile.Seek(0, io.SeekStart)
		if err != nil {
			log.Println("can't rewind archive file pointer to the beginning:", err)
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}
		var gzr *gzip.Reader
		gzr, err = gzip.NewReader(archiveFile)
		if err != nil {
			log.Printf("Error reading from %s: %s", url, err)
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}
		tarReader := tar.NewReader(gzr)
		var out *os.File
		out, err = os.Create(dbPath + ".tmp")
		if err != nil {
			log.Println("error opening "+dbPath, ":", err)
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			return
		}
		log.Println("Unpacking tar archive with geo database...")
		for {
			var header *tar.Header
			header, err = tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error reading from tar archive: %s", err)
				out.Close()
				archiveFile.Close()
				_ = os.Remove(dbPath + ".tar.gz")
				_ = os.Remove(dbPath + ".tmp")
				return
			}
			if header.Typeflag != tar.TypeReg {
				continue
			}
			if !strings.HasSuffix(header.Name, ".mmdb") {
				continue
			}
			_, err = io.Copy(out, tarReader)
			if err != nil {
				log.Println("can't extract to", dbPath, ":", err)
				out.Close()
				archiveFile.Close()
				_ = os.Remove(dbPath + ".tar.gz")
				_ = os.Remove(dbPath + ".tmp")
				return
			}
			archiveFile.Close()
			_ = os.Remove(dbPath + ".tar.gz")
			g.Lock()
			// First closing db reader
			if g.reader != nil {
				_ = g.reader.Close()
				g.reader = nil
			}
			// Moving tmp file to the dbPath
			var realOut *os.File
			realOut, err = os.Create(dbPath)
			if err != nil {
				g.Unlock()
				out.Close()
				log.Println("Can't create file " + dbPath)
				_ = os.Remove(dbPath + ".tmp")
				return
			}
			_, _ = out.Seek(0, io.SeekStart)
			_, err = io.Copy(realOut, out)
			out.Close()
			realOut.Close()
			_ = os.Remove(dbPath + ".tmp")
			// Opening db
			reader, err := geoip2.Open(dbPath)
			if err != nil {
				log.Printf("can't open maxmind database: %s", err)
				g.Unlock()
				return
			} else {
				g.reader = reader
				log.Printf("successfully loaded %s", filepath.Base(dbPath))
			}
			g.Unlock()
			return
		}
	}
}

func InitGeoIP(Path string, geoUrl string) {
	loadGeoDB(Path, geoUrl)
	go func() {
		for range time.Tick(time.Hour+time.Duration(rand.Intn(1800))*time.Second) {
			loadGeoDB(Path, geoUrl)
		}
	}()
}

func (g *MaxmindGeoIP2) Country(ip net.IP) (string, error) {
	g.Lock()
	defer g.Unlock()
	var reader *geoip2.Reader
	reader = g.reader
	if reader == nil {
		return "", errors.New("maxmind not initialized")
	}
	c, err := reader.Country(ip)
	if err != nil {
		return "", errors.New("maxmind error: " + err.Error())
	}
	return c.Country.IsoCode, nil
}

func Country(ip net.IP) (string, error) {
	return geoIP.Country(ip)
}

func ExitCleanup() {
	done := make(chan struct{}, 1)
	go func() {
		for {
			geoIP.Lock()
			if geoIP.reloading {
				time.Sleep(time.Millisecond * 5)
				geoIP.Unlock()
				continue
			}
			geoIP.Unlock()
			break
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(time.Second * 15):
		log.Println("GeoIP reloading wait timed out!")
	}
}
