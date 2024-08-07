require 'socket'
require 'openssl'

server_private_key = OpenSSL::PKey::RSA.new(File.read('server_private_key.pem'))
listener_public_key = OpenSSL::PKey::RSA.new(File.read('listener_public_key.pem'))

server_ip = '0.tcp.sa.ngrok.io'
server_port = 19968

aes_key = '12345678901234567890123456789012'
aes_iv = '0123456789abcdef'
def aes_encrypt(data, key, iv)
  cipher = OpenSSL::Cipher.new('aes-256-cbc')
  cipher.encrypt
  cipher.key = key
  cipher.iv = iv

  encrypted = cipher.update(data) + cipher.final

  return encrypted
end

def aes_decrypt(data, key, iv)
  cipher = OpenSSL::Cipher.new('aes-256-cbc')
  cipher.decrypt
  cipher.key = key
  cipher.iv = iv

  decrypted = cipher.update(data) + cipher.final

  return decrypted
end

puts "Client connecting to the server..."

socket = TCPSocket.new(server_ip, server_port)

puts "Connected to the server."

# Interactive shell loop
loop do
  print "> "
  command = gets.chomp
  encrypted_command = listener_public_key.public_encrypt(command)
  encrypted_command_2 = aes_encrypt(encrypted_command, aes_key, aes_iv)

  socket.puts encrypted_command_2
  socket.puts
  socket.flush

  puts "Command sent"

  encrypted_response = socket.gets("\n\n").strip
  #puts "Received encrypted response: #{encrypted_response}"

  decrypted_response = aes_decrypt(encrypted_response, aes_key, aes_iv)
  puts "Decrypted response: #{decrypted_response}"
end

socket.close
