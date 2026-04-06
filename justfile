set script-interpreter := ['bash']

mod database
mod prod
mod test
mod misc
mod new

@default: (misc::list 'database' 'prod' 'test' 'misc' 'new')

[doc('pushes and fetches the code in the server, risky')]
[group('unstable')]
deploy:
    -git add .
    -git commit
    -git push origin wip
    ssh mytempogps@mytempogps.com.br \
        just -f /home/mytempogps/pvauto/pvapi/justfile prod::new
