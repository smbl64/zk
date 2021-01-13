package sqlite

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestNoteDAOIndexed(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		expected := []paths.Metadata{
			{
				Path:     "f39c8.md",
				Modified: time.Date(2020, 1, 20, 8, 52, 42, 0, time.UTC),
			},
			{
				Path:     "index.md",
				Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
			},
			{
				Path:     "log/2021-01-03.md",
				Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
			},
			{
				Path:     "log/2021-01-04.md",
				Modified: time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "log/2021-02-04.md",
				Modified: time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "ref/test/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
			{
				Path:     "ref/test/b.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
		}

		c, err := dao.Indexed()
		assert.Nil(t, err)

		actual := make([]paths.Metadata, 0)
		for a := range c {
			actual = append(actual, a)
		}
		assert.Equal(t, actual, expected)
	})
}

func TestNoteDAOAdd(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Add(note.Metadata{
			Path:      "log/added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:  time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
			Checksum:  "check",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `path = "log/added.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:      "log/added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Checksum:  "check",
			Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:  time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
		})
	})
}

// Check that we can't add a duplicate note with an existing path.
func TestNoteDAOAddExistingNote(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Add(note.Metadata{Path: "ref/test/a.md"})
		assert.Err(t, err, "ref/test/a.md: can't add note to the index: UNIQUE constraint failed: notes.path")
	})
}

func TestNoteDAOUpdate(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Update(note.Metadata{
			Path:      "ref/test/a.md",
			Title:     "Updated note",
			Body:      "Updated body",
			Checksum:  "updated checksum",
			WordCount: 42,
			Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:  time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `path = "ref/test/a.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:      "ref/test/a.md",
			Title:     "Updated note",
			Body:      "Updated body",
			Checksum:  "updated checksum",
			WordCount: 42,
			Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:  time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
		})
	})
}

func TestNoteDAOUpdateUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Update(note.Metadata{
			Path: "unknown/unknown.md",
		})
		assert.Err(t, err, "unknown/unknown.md: failed to update note index: note not found in the index")
	})
}

func TestNoteDAORemove(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Remove("ref/test/a.md")
		assert.Nil(t, err)
	})
}

func TestNoteDAORemoveUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Remove("unknown/unknown.md")
		assert.Err(t, err, "unknown/unknown.md: failed to remove note index: note not found in the index")
	})
}

func TestNoteDAOFindAll(t *testing.T) {
	testNoteDAOFindPaths(t, note.FinderOpts{}, []string{
		"ref/test/b.md",
		"f39c8.md",
		"ref/test/a.md",
		"log/2021-02-04.md",
		"index.md",
		"log/2021-01-03.md",
		"log/2021-01-04.md",
	})
}

func TestNoteDAOFindLimit(t *testing.T) {
	testNoteDAOFindPaths(t, note.FinderOpts{Limit: 2}, []string{
		"ref/test/b.md",
		"f39c8.md",
	})
}

func TestNoteDAOFindMatch(t *testing.T) {
	testNoteDAOFind(t,
		note.FinderOpts{
			Filters: []note.Filter{note.MatchFilter("daily | index")},
		},
		[]note.Match{
			{
				Snippet: "<zk:match>Index</zk:match> of the Zettelkasten",
				Metadata: note.Metadata{
					Path:      "index.md",
					Title:     "Index",
					Body:      "Index of the Zettelkasten",
					WordCount: 4,
					Created:   time.Date(2019, 12, 4, 11, 59, 11, 0, time.UTC),
					Modified:  time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
					Checksum:  "iaefhv",
				},
			},
			{
				Snippet: "A <zk:match>daily</zk:match> note",
				Metadata: note.Metadata{
					Path:      "log/2021-01-03.md",
					Title:     "January 3, 2021",
					Body:      "A daily note",
					WordCount: 3,
					Created:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum:  "qwfpgj",
				},
			},
			{
				Snippet: "A second <zk:match>daily</zk:match> note",
				Metadata: note.Metadata{
					Path:      "log/2021-01-04.md",
					Title:     "January 4, 2021",
					Body:      "A second daily note",
					WordCount: 4,
					Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Checksum:  "arstde",
				},
			},
			{
				Snippet: "A third <zk:match>daily</zk:match> note",
				Metadata: note.Metadata{
					Path:      "log/2021-02-04.md",
					Title:     "February 4, 2021",
					Body:      "A third daily note",
					WordCount: 4,
					Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:  time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
					Checksum:  "earkte",
				},
			},
		},
	)
}

func TestNoteDAOFindInPath(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.PathFilter([]string{"log/2021-01-*"})},
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindInMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.PathFilter([]string{"ref", "index.md"})},
		},
		[]string{"ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingPath(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.ExcludePathFilter([]string{"log"})},
		},
		[]string{"ref/test/b.md", "f39c8.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.ExcludePathFilter([]string{"ref", "log/2021-01-*"})},
		},
		[]string{"f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindCreatedOn(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 10, 12, 45, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateOn,
				},
			},
		},
		[]string{"log/2021-01-03.md"},
	)
}

func TestNoteDAOFindCreatedBefore(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2019, 12, 04, 11, 59, 11, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateBefore,
				},
			},
		},
		[]string{"ref/test/b.md", "ref/test/a.md"},
	)
}

func TestNoteDAOFindCreatedAfter(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateAfter,
				},
			},
		},
		[]string{"log/2021-02-04.md", "log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindModifiedOn(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 01, 20, 10, 12, 45, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateOn,
				},
			},
		},
		[]string{"f39c8.md"},
	)
}

func TestNoteDAOFindModifiedBefore(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 01, 20, 8, 52, 42, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateBefore,
				},
			},
		},
		[]string{"ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindModifiedAfter(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateAfter,
				},
			},
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func testNoteDAOFindPaths(t *testing.T, opts note.FinderOpts, expected []string) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		actual := make([]string, 0)
		count, err := dao.Find(opts, func(m note.Match) error {
			actual = append(actual, m.Path)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, count, len(expected))
		assert.Equal(t, actual, expected)
	})
}

func testNoteDAOFind(t *testing.T, opts note.FinderOpts, expected []note.Match) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		actual := make([]note.Match, 0)
		count, err := dao.Find(opts, func(m note.Match) error {
			actual = append(actual, m)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, count, len(expected))
		assert.Equal(t, actual, expected)
	})
}

func testNoteDAO(t *testing.T, callback func(tx Transaction, dao *NoteDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewNoteDAO(tx, &util.NullLogger))
	})
}

type noteRow struct {
	Path, Title, Body, Checksum string
	WordCount                   int
	Created, Modified           time.Time
}

func queryNoteRow(tx Transaction, where string) (noteRow, error) {
	var row noteRow
	err := tx.QueryRow(fmt.Sprintf(`
		SELECT path, title, body, word_count, checksum, created, modified
		  FROM notes
		 WHERE %v
	`, where)).Scan(&row.Path, &row.Title, &row.Body, &row.WordCount, &row.Checksum, &row.Created, &row.Modified)
	return row, err
}