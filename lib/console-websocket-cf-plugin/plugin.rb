require 'console-websocket-cf-plugin'
require 'eventmachine'
require 'termios'
require 'stringio'
require 'base64'
require 'json'
require 'zlib'
require 'faye/websocket'

module ConsoleWebsocketCfPlugin

  include EM::Protocols::LineText2

  class KeyboardHandler < EM::Connection
    
    include EM::Protocols::LineText2

    def initialize(ws, guid)
      @ws = ws
      @guid = guid
      @buffer = ''
    end

    def receive_line(data)
      EM.stop if data == "exit"
      @ws.lines_in << data
      @ws.send data
    end

    def move_history(direction)
      puts direction
    end

  end
    
  class Plugin < CF::CLI
    
    def precondition
      # skip all default preconditions
    end

    desc "Open a console to an application container"
    group :admin
    input :app, :desc => "Application to connect to", :argument => true,
          :from_given => by_name(:app)
    input :instance, :desc => "Instance (index) to connect", :default => 0

    def console

      app = input[:app]

      app_version = app.manifest[:entity][:version]
      guid = "#{app_version}/#{input[:instance]}"
      ws_url = "wss://#{app.url}:4443/#{guid}"

      puts "Starting connection to #{ws_url}"

      start_connection(app, guid, ws_url)

    end

    private

    def start_connection(app, guid, ws_url)
      
      EM.run {

        Faye::WebSocket::Client.class_eval <<-BYTES_IN
          attr_accessor :lines_in
        BYTES_IN

        ws = Faye::WebSocket::Client.new(ws_url, nil, { :headers => { 'Origin' => "http://#{app.url}:4443" }})
        ws.lines_in = []

        ws.on :error do |event|
          p [:error]
          p ws.status
        end

        ws.on :message do |event|
          msg = event.data

          msg = msg[(ws.lines_in.last.length)..(msg.length - 1)].lstrip if msg.start_with?(ws.lines_in.last)
          print msg
        end

        ws.on :close do |event|
          ws = nil
          EM.stop
        end
        
        EM.open_keyboard(KeyboardHandler, ws, guid)
      }

    end

  end
end