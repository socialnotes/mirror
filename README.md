# Mirror
This is the software that runs [socialnotes.eu](https://socialnotes.eu/) .

This program aims to improve the university life experience,
providing a simple an intuitive website,
dedicated to the students, 
to share notes and teaching materials.
It want to be the reference _loco_ for every kind of students,
being looking for old exams files, or willing to share their knowledge via notes or just helping.

Even tough a forum is not integrated, the discussion is currently taken on FB groups.

The project is currently configured for the University of Trento:
As can be seen [here](https://github.com/socialnotes/mirror/blob/8d62da77c534f32d9e2889ed7bcda315ee667e9f/views/upload.go#L51)
it only accepts uploads from users provided with a valid "unitn" email.
This was necessary to prevent the system from becoming a general file sharing service (with all the related issue).

We are just students working for free for all the other students.
This project is __not__ affiliated in any way with the University of Trento.


## HOW CAN I CONTRIBUTE ?
- Upload your notes

    By design you email will not be shown next to the files you upload;
    to encourage the sharing even of not perfectly styled notes.<br>
    If you wish you can obviously sign them.
- Report or Fix issues [here](https://github.com/socialnotes/mirror/issues)
- Propose or Implement features [here](https://github.com/socialnotes/mirror/issues)

## REQUIREMENTS:
- go 1.5+

## INSTALL:
- Install the software as `go get github.com/socialnotes/mirror`
- Install the indexer as `go get github.com/socialnotes/mirror/indexer`
- Build the index as `indexer -base-dir /srv/files/ -db-file /srv/db.bolt -email admin@example.com`
- Run as `mirror -base-dir /srv/files/ -db-file /srv/db.bolt -mailgun-api-key <api-key> -mailgun-domain <api-domain>`

## OTHERS:
[Authors](AUTHORS.md) & License: [MIT](LICENSE.md)
