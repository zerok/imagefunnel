# ImageFunnel: Continuous image resizing

This is a little tool checks a given S3 compatible bucket or a folder for new
images and resizes them according to a set of provided profiles.


## Usage

The following command will check all files in the current working directory
(not recursively!) against the provided funnel profile and generate the missing
ones:

```
$ imagefunnel  --profile profile.yaml .
```

The profile to be used should resize all JPEG files (unless they contain the
string `large`) into images 500px wide if they are portrait or 1000px if they
are landscape:

```
source:
  include:
    - ".*\\.jpe?g"
  exclude:
    - ".*large.*"
target:
  filename: "{{ .Stem }}.large.{{ .Ext }}"
  portrait_size: "500x"
  landscape_size: "1000x"
```

