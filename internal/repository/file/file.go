package file

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
)

type FileBackup struct {
	file *os.File
}

func NewFileBackup(path string) FileBackup {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error().Err(err).Msg("failed to open backup file")
		file = os.Stdout
	}
	return FileBackup{file: file}
}

func (fs FileStorage) Add(ctx context.Context, metric metrics.Metric) error {
	if err := fs.storage.Add(ctx, metric); err != nil {
		log.Error().Err(err).Msg("failed to add metric")
		return err
	}

	if err := fs.Save(ctx); err != nil {
		log.Error().Err(err).Msg("backup failed")
		return err
	}
	return nil
}

func (fs FileStorage) Get(ctx context.Context, mType, name string) (metrics.Metric, error) {
	metric, err := fs.storage.Get(ctx, mType, name)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get metric type:%s, name:%s", mType, name)
		return metrics.Metric{}, err
	}
	return metric, nil
}

func (fs FileStorage) List(ctx context.Context) []metrics.Metric {
	return fs.storage.List(ctx)
}

func (fs FileStorage) Close() error {
	return errors.Join(fs.storage.Close(), fs.file.Close())
}

func (fs FileStorage) Load(_ context.Context) ([]byte, error) {
	istreamInfo, err := fs.file.Stat()
	if err != nil {
		log.Error().Err(err).Msg("error getting istream info")
		return make([]byte, 0), err
	}

	data := make([]byte, istreamInfo.Size())
	_, err = fb.file.Read(data)
	if err != nil {
		log.Error().Err(err).Msg("error reading from istream")
		return make([]byte, 0), err
	}
	return data, nil
}

func (fb FileBackup) Save(_ context.Context, data []byte) error {
	if _, err := fb.file.Seek(0, 0); err != nil {
		log.Error().Err(err).Msgf("moving file prt to begining %v", fb.file)
		return err
	}

	if _, err := fb.file.Write(append(data, '\n')); err != nil {
		log.Error().Err(err).Msg("error writing data to ostream")
		return err
	}
	return nil
}

func (fb FileBackup) Close() error {
	if fb.file == os.Stdout || fb.file == os.Stderr {
		return nil
	}
	return fb.file.Close()
}
