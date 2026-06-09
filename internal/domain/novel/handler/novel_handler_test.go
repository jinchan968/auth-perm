package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadLimitedRejectsOversizedContent(t *testing.T) {
	content, err := readLimited(strings.NewReader("123456"), 5, "文件过大")

	require.Error(t, err)
	require.Nil(t, content)
	require.Contains(t, err.Error(), "文件过大")
}

func TestReadLimitedAllowsContentAtLimit(t *testing.T) {
	content, err := readLimited(strings.NewReader("12345"), 5, "文件过大")

	require.NoError(t, err)
	require.Equal(t, "12345", string(content))
}

func TestReadMarkdownBundleZipRejectsTooManyFiles(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for i := 0; i <= maxMarkdownBundleFiles; i++ {
		file, err := writer.Create(fmt.Sprintf("%04d.md", i))
		require.NoError(t, err)
		_, err = file.Write([]byte("# 标题\n\n正文"))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	files, err := readMarkdownBundleZip(archive.Bytes())

	require.Error(t, err)
	require.Nil(t, files)
	require.Contains(t, err.Error(), "文件数量")
}

func TestReadMarkdownBundleZipRejectsExpandedOversize(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("oversize.md")
	require.NoError(t, err)
	_, err = file.Write([]byte(strings.Repeat("a", int(maxMarkdownFileBytes)+1)))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	files, err := readMarkdownBundleZip(archive.Bytes())

	require.Error(t, err)
	require.Nil(t, files)
	require.Contains(t, err.Error(), "单个 Markdown")
}
