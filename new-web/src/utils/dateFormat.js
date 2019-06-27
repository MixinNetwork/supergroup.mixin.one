import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

export function fromNow(t) {
  return dayjs(t).fromNow()
}

export function format(t) {
  return dayjs(t).format('YYYY-MM-DD HH:mm:ss')
}