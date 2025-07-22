package metricsstorage

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

func WriteMetricsStorage(ostream *os.File) error {
	if _, err := ostream.Seek(0, 0); err != nil {
		return err
	}

	data, err := MS.MarshalJSON()
	if err != nil {
		log.Error().Err(err)
		return err
	}

	if _, err := ostream.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func RunSavingStorageRoutine(ostream *os.File, interval int) {
	go func() {
		for {
			if err := WriteMetricsStorage(ostream); err != nil {
				log.Fatal().Err(err)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

func ReadMetricsStorage(istream *os.File) error {
	istreamInfo, err := istream.Stat()
	if err != nil {
		log.Error().Err(err)
		return err
	}

	data := make([]byte, istreamInfo.Size())
	_, err = istream.Read(data)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	if err := MS.UnmarshalJSON(data); err != nil {
		log.Error().Err(err)
		return err
	}
	return nil
}
