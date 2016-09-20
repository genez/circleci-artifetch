# circleci-artifetch [![Build Status](https://travis-ci.org/genez/circleci-artifetch.svg?branch=master)](https://travis-ci.org/genez/circleci-artifetch)
Artifact downloader for CircleCI Continuous Integration

CircleCI has a nice REST API that can be used to retrieve artifacts.  
Usage is straight-forward:

1) set the CIRCLE_CI_TOKEN environment variable with a valid token (see https://circleci.com/docs/api/)  
2) launch circleci-artifetch for example: `circleci-artifetch -vcs=bitbucket -user=yourusername -project=yourproject`  
3) check extended usage info `circleci-artifetch -h` for optional parameters  
