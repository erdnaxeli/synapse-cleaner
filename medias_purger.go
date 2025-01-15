package synapsecleaner

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
)

type media struct {
	id    string
	local bool
}

func (m media) Path(root string) string {
	var directory string
	if m.local {
		directory = "local_content"
	} else {
		directory = "remote_content"
	}

	return filepath.Join(root, directory, m.id[:2], m.id[2:4], m.id[4:])
}

type MediasPurger struct {
	DatabaseUri     string
	MediasDirectory string
}

func (mp MediasPurger) Run() error {
	medias, err := mp.listMedias()
	if err != nil {
		return err
	}

	usedMedias, err := mp.listUsedMedias()
	if err != nil {
		return err
	}

	fmt.Printf("%d used medias found, for %d total medias.\n", len(usedMedias), len(medias))

	if len(usedMedias) == 0 {
		fmt.Println("There is no used medias, this script refuse to delete all existing medias, stopping there.")
		return nil
	}

	unusedMedias := mp.getUnusedMedias(medias, usedMedias)
	if len(unusedMedias) == 0 {
		fmt.Println("No unused medias, stopping there.")
		return nil
	}

	err = mp.deleteMedias(unusedMedias)
	if err != nil {
		return err
	}

	return nil
}

func (mp MediasPurger) listMedias() ([]media, error) {
	var medias []media

	localContentPath := filepath.Join(mp.MediasDirectory, "local_content")
	localMedias, err := mp.listMediasPath(localContentPath)
	for mediaId := range localMedias {
		medias = append(medias, media{
			id:    mediaId,
			local: true,
		})
	}
	if *err != nil {
		return nil, *err
	}

	/*
		remoteContentPath := filepath.Join(mp.MediasDirectory, "remote_content")
		remoteMedias, err := mp.listMediasPath(remoteContentPath)
		for mediaId := range remoteMedias {
			medias = append(medias, media{
				id:    mediaId,
				local: false,
			})
		}
		if *err != nil {
			return nil, *err
		}
	*/

	return medias, nil
}

func (mp MediasPurger) listMediasPath(path string) (iter.Seq[string], *error) {
	var returnedErr error

	f := func(yield func(string) bool) {
		// The media are in <path>/aa/bb/cccccccccccccccccccc
		entries, err := os.ReadDir(path)
		if err != nil {
			returnedErr = err
			return
		}

		for _, entryPart1 := range entries {
			if !entryPart1.IsDir() {
				continue
			}

			part1 := entryPart1.Name()

			part1Path := filepath.Join(path, part1)
			entriesPart2, err := os.ReadDir(part1Path)
			if err != nil {
				returnedErr = err
				return
			}

			for _, entryPart2 := range entriesPart2 {
				if !entryPart2.IsDir() {
					continue
				}

				part2 := entryPart2.Name()

				part2Path := filepath.Join(part1Path, part2)
				entriesPart3, err := os.ReadDir(part2Path)
				if err != nil {
					returnedErr = err
					return
				}

				for _, entryPart3 := range entriesPart3 {
					if entryPart3.IsDir() {
						continue
					}

					part3 := entryPart3.Name()
					entryPart3.Info()

					mediaId := fmt.Sprintf("%s%s%s", part1, part2, part3)
					if !yield(mediaId) {
						return
					}
				}
			}
		}
	}

	return f, &returnedErr
}

func (mp MediasPurger) listUsedMedias() ([]media, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, mp.DatabaseUri)
	if err != nil {
		return nil, err
	}

	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "select media_id from local_media_repository")
	if err != nil {
		return nil, err
	}

	var medias []media
	for rows.Next() {
		var mediaId string
		err := rows.Scan(&mediaId)
		if err != nil {
			return nil, err
		}

		medias = append(medias, media{id: mediaId, local: true})
	}

	return medias, nil
}

func (mp MediasPurger) getUnusedMedias(medias []media, usedMedias []media) []media {
	return DiffSlices(medias, usedMedias)
}

func (mp MediasPurger) deleteMedias(medias []media) error {
	fmt.Println("Deleting medias...")
	for _, media := range medias {
		fmt.Printf("%s -> %s\n", media.Path(mp.MediasDirectory), media.id)
	}

	return nil
}
