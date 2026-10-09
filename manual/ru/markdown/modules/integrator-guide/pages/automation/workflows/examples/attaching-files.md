# Attaching Files to Records
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

Attaching files to records is straight forward but it requires a few steps.

!!! important
    This example receives files via integration gateway.


.In this example we will:
- Create a new record,
- upload an attachment,
- attach it to the newly created record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/workflow-base.png",
    "alias": "automation-workflows-examples-attaching-files-workflow-base",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 131,
    "y": 160,
    "w": 1186,
    "h": 394
  },
  "focus": {
    "x": 165,
    "y": 338,
    "w": 1103,
    "h": 173
  },
  "annotations": []
}

Firstly, we need to prepare a new record using the "Compose Record Maker" function.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/make-record.png",
    "alias": "automation-workflows-examples-attaching-files-make-record",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 259,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Next, we need to upload the file and prepare an `Attachment`.
The attachment needs to specify the used module field name as well as the record we've prepared earlier.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/upload-attachment.png",
    "alias": "automation-workflows-examples-attaching-files-upload-attachment",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 510,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Next, we need to reference the Attachment in the newly prepared record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/add-reference.png",
    "alias": "automation-workflows-examples-attaching-files-add-reference",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 763,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

Lastly, we create the record.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/save-record.png",
    "alias": "automation-workflows-examples-attaching-files-save-record",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 1014,
    "y": 388,
    "w": 180,
    "h": 73
  }]
}

The newly created record can be seen on the record list

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/attaching-files/record-list.png",
    "alias": "automation-workflows-examples-attaching-files-record-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 346,
    "y": 90,
    "w": 1538,
    "h": 410
  },
  "annotations": []
}
