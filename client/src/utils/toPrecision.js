export default (n, c) => {
  return (Math.floor(Number(n) * (10 ** c)) / (10 ** c)).toFixed(c)
}