# Alpine Packages Diff

This tool detects the differences between two Alpine APKINDEX files and prints out the deltas.

## Examples

Using two APKINDEX.tar.gz files
```bash
$ ./alpine-package-diff -new NEW_APKINDEX.tar.gz -old OLD_APKINDEX.tar.gz -showAdded
```

Using one repo directory and one file
```bash
./yum-package-diff -new output/ -old OLD_APKINDEX.tar.gz -showAdded -output filelist.txt
```

Using just a new file, this gives you a full list
```bash
$ ./alpine-package-diff -new NEW_APKINDEX.tar.gz -old "" -showAdded
```

## Usage
```bash
$ ./alpine-package-diff -h
Alpine Index Diff, Version: 0.1.20220317.1355

Usage: ./alpine-package-diff [options...]

  -new string
        The newer APKINDEX.tar.gz file or repodata/ dir for comparison (default "NEW_APKINDEX.tar.gz")
  -old string
        The older APKINDEX.tar.gz file or repodata/ dir for comparison (default "OLD_APKINDEX.tar.gz")
  -output string
        Output for comparison result (default "-")
  -repo string
        Repo path to use in file list (default "/latest-stable/main/x86_64")
  -showAdded
        Display packages only in the new list
  -showCommon
        Display packages in both the new and old lists
  -showRemoved
        Display packages only in the old list
```
