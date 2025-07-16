import base64
from Crypto.Cipher import AES

def pad_key(mobile):
    return mobile.ljust(32, '\0').encode('utf-8')

def pkcs7_unpad(data):
    pad_len = data[-1]
    return data[:-pad_len]

def decrypt_user_data(user_data_b64, mobile_no):
    key = pad_key(mobile_no)
    iv = b'\x00' * 16
    encrypted = base64.b64decode(user_data_b64)
    cipher = AES.new(key, AES.MODE_CBC, iv)
    decrypted = cipher.decrypt(encrypted)
    unpadded = pkcs7_unpad(decrypted)
    return unpadded.decode('utf-8')

# Example usage:
user_data_b64 = "0j417EmZvxMWHNNUJlJFdtg+p97srb5nuy0MjW41A+dp6p57wxe82ZkGIoLcNNtG"
mobile_no = "5985899652"

decrypted_json = decrypt_user_data(user_data_b64, mobile_no)
print("Decrypted JSON:", decrypted_json)

# Encrypt and then decrypt in the same script
# from Crypto.Cipher import AES
# import base64

# def pad_key(mobile):
#     return mobile.ljust(32, '\0').encode('utf-8')

# def pkcs7_pad(data):
#     pad_len = 16 - (len(data) % 16)
#     return data + bytes([pad_len] * pad_len)

# def pkcs7_unpad(data):
#     pad_len = data[-1]
#     return data[:-pad_len]

# def encrypt_user_data(payload_dict, mobile_no):
#     key = pad_key(mobile_no)
#     iv = b'\x00' * 16
#     json_bytes = json.dumps(payload_dict, separators=(',', ':')).encode('utf-8')
#     padded = pkcs7_pad(json_bytes)
#     cipher = AES.new(key, AES.MODE_CBC, iv)
#     encrypted = cipher.encrypt(padded)
#     return base64.b64encode(encrypted).decode('utf-8')

# def decrypt_user_data(user_data_b64, mobile_no):
#     key = pad_key(mobile_no)
#     iv = b'\x00' * 16
#     encrypted = base64.b64decode(user_data_b64)
#     cipher = AES.new(key, AES.MODE_CBC, iv)
#     decrypted = cipher.decrypt(encrypted)
#     unpadded = pkcs7_unpad(decrypted)
#     return unpadded.decode('utf-8')

# import json
# mobile_no = "5985899652"
# payload = {"mobile_no": mobile_no, "device_id": "abc123"}
# user_data_b64 = encrypt_user_data(payload, mobile_no)
# print("Encrypted user_data:", user_data_b64)
# decrypted_json = decrypt_user_data(user_data_b64, mobile_no)
# print("Decrypted JSON:", decrypted_json)