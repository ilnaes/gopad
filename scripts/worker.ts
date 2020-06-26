// handles websocket messaging with server

onmessage = (e) => {
  let message = e.data
}

const sleep = (milliseconds: number) => {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}
