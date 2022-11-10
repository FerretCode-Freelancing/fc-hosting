require "roda"

module LogHTTPRouter
  http_response = [200, {
    'Content-Type' => 'application/json',
    'Content-Length' => 32
  }, ['Please connect using websockets.']]

  ws_response = [0, {}, []].freeze

  def self.call env
    if(env['rack.upgrade?'.freeze] == :websocket)
      env['rack.upgrade'.freeze] = WebsocketConnector

      return ws_response
    end

    http_response
  end
end

module WebsocketConnector

end
