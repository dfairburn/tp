1. Create a `new` subcommand to fill out templates quickly using fzf where applicable. General flow:
    * `tp new`
    * get url from user
    * fzf method from `GET, PUT, DELETE, etc...`
    * fzf headers
    * get body (open vim? or option to have none)


2. Create a `cp` subcommand to copy templates. Again use fzf here


3. Create an `edit` subcommand so you can find and edit teplates from the cli. Fzf on all templates then open in vim


4. Would be nice to have the idea of folders, or labels? Maybe an extra field in the template
