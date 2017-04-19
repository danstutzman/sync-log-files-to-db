var child_process = require('child_process')

exports.handler = function(event, context) {
  var proc = child_process.spawn('./test', [ JSON.stringify(event) ])
  proc.stdout.on('data', function(data) {
    console.error(data.toString().trim())
  })
  proc.stderr.on('data', function(data) {
    console.log(data.toString().trim())
  })
  proc.on('error', function(err) {
    console.error('error from child', err)
  })
  proc.on('close', function(code){
    if(code !== 0) {
      return context.done(new Error("Process exited with non-zero status code"))
    }
    context.done(null)
  })
}
