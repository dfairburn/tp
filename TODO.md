1. Create a `new` subcommand to fill out templates quickly using fzf where applicable. General flow:
    * `tp new`
    * get url from user
    * fzf method from `GET, PUT, DELETE, etc...`
    * fzf headers
    * get body (open vim? or option to have none)

2. Create a `cp` subcommand to copy templates. Again use fzf here

3. Create an `edit` subcommand so you can find and edit teplates from the cli. Fzf on all templates then open in vim

4. Would be nice to have the idea of folders, or labels? Maybe an extra field in the template

5. Might be good to base this off of jinja templating...? Instead of having my own janky template language based off of the go templating thingy?

6. Would be good to have all the organizational and editorial stuff in a library, maybe separate to this repo?
		* this repo could then focus on the whole cli stuff and being specific over http clients
