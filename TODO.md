1. Create a `new` subcommand to fill out templates quickly using fzf where applicable. General flow:
    * `tp new`
    * get url from user
    * fzf method from `GET, PUT, DELETE, etc...`
    * fzf headers
    * get body (open vim? or option to have none)

2. Create a `cp` subcommand to copy templates. Again use fzf here

3. Create an `edit` subcommand, so you can find and edit templates from the cli. Fzf on all templates then open in vim

4. Would be nice to have the idea of folders, or labels? Maybe an extra field in the template

5. Might be good to base this off of jinja templating...? Instead of having my own janky template language based off of the go templating thingy?

6. Would be good to have all the organizational and editorial stuff in a library, maybe separate to this repo?
		* this repo could then focus on the whole cli stuff and being specific over http clients

Hackday - 26/04/2024

- PP output for html and json
  - maybe this can just be done by the unix way of feeding the output to `jq`?
- Implement `get-token` into `use`
  - This can also be done by just using an override, so using -o token:$(get-token). Makes it a bit more versatile
- Suggestions of vars that are available
- cmd alias to edit vars file
- cmd alias to edit config file
- shell autocompletion on overrides
  - This is pretty difficult. Ideally I'd like to know all the variables that are used in both the .vars file and those
    that aren't defined but reside in templates. I don't think this is realistically possible, as we'd have to run that
    `tp completion zsh` when we have all our templates defined. Which might be okay if we ship some default templates,
    but not great when we want to add stuff... Maybe we can regenerate templates each time we call `tp open` or something?
- vim autocompletion on overrides
