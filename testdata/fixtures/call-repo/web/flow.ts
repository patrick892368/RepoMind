export function createOrder(input: unknown) {
  validateOrder(input);
  saveOrder(input);
  notifyUser(input);
}

function validateOrder(input: unknown) {
  return input;
}

function saveOrder(input: unknown) {
  writeAudit(input);
}

function notifyUser(input: unknown) {
  writeAudit(input);
}

function writeAudit(input: unknown) {
  return input;
}
