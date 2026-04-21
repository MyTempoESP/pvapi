set script-interpreter := ['bash']

import? 'gemini.just'

mod database
mod prod
mod test
mod misc
mod new
mod ide

@default: (misc::list 'database' 'prod' 'test' 'misc' 'new' 'ide')

