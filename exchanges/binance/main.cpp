#include <boost/beast/core.hpp>
#include <boost/beast/websocket.hpp>
#include <boost/beast/websocket/ssl.hpp>
#include <boost/asio/connect.hpp>
#include <boost/asio/ip/tcp.hpp>
#include <boost/asio/ssl.hpp>
#include <iostream>
#include <string>

namespace beast = boost::beast;
namespace websocket = beast::websocket;
namespace net = boost::asio;
namespace ssl = net::ssl;
using tcp = net::ip::tcp;

int main() {
    try {
        net::io_context ioc;
        ssl::context ctx{ssl::context::tlsv12_client};

        tcp::resolver resolver{ioc};
        websocket::stream<ssl::stream<tcp::socket>> ws{ioc, ctx};

        auto const results = resolver.resolve("stream.binance.com", "9443");
        auto ep = net::connect(beast::get_lowest_layer(ws), results);

        if(!SSL_set_tlsext_host_name(ws.next_layer().native_handle(), "stream.binance.com")) {
            throw beast::system_error{
                beast::error_code{static_cast<int>(::ERR_get_error()), net::error::get_ssl_category()},
                "Failed to set SNI"
            };
        }

        ws.next_layer().handshake(ssl::stream_base::client);
        ws.set_option(websocket::stream_base::decorator(
            [](websocket::request_type& req) {
                req.set(beast::http::field::user_agent, "binance-connector");
            }
        ));

        ws.handshake("stream.binance.com", "/ws/btcusdt@kline_1m");

        std::cout << "Connected to Binance WebSocket (BTCUSDT 1m klines)" << std::endl;

        while(true) {
            beast::flat_buffer buffer;
            ws.read(buffer);
            std::cout << beast::make_printable(buffer.data()) << std::endl;
        }

        ws.close(websocket::close_code::normal);

    } catch(std::exception const& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return EXIT_FAILURE;
    }
    return EXIT_SUCCESS;
}
