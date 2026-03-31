set script-interpreter := ['bash']

mod database
mod prod
mod test
mod misc
mod new

@default: (misc::list 'database' 'prod' 'test' 'misc' 'new')
