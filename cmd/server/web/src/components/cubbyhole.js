const cubbyholeKey = "oauth:cubbyhole:key"

export default function cubbyhole(secret) {
    if (!localStorage) {
        return ["", new Error("localStorage not found, but is required")]
    }

    let storedSecret = localStorage.getItem(cubbyholeKey)
    if (secret.length > 0) {
        localStorage.setItem(cubbyholeKey, secret)
        storedSecret = secret
    }

    if (!storedSecret) {
        return ["", new Error("encryption key is missing")]
    }

    return [storedSecret, null]
}
