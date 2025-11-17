#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/websocket.hpp>
#include <boost/beast/websocket/ssl.hpp>
#include <boost/asio/connect.hpp>
#include <boost/asio/ip/tcp.hpp>
#include <boost/asio/ssl.hpp>
#include <boost/json.hpp>
#include <iostream>
#include <string>

namespace beast = boost::beast;
namespace http = beast::http;
namespace websocket = beast::websocket;
namespace net = boost::asio;
namespace ssl = net::ssl;
namespace json = boost::json;
using tcp = net::ip::tcp;

std::string get_kucoin_token(net::io_context& ioc, ssl::context& ctx) {
    tcp::resolver resolver{ioc};
    ssl::stream<tcp::socket> stream{ioc, ctx};

    auto const results = resolver.resolve("api.kucoin.com", "443");
    net::connect(stream.next_layer(), results);

    if(!SSL_set_tlsext_host_name(stream.native_handle(), "api.kucoin.com")) {
        throw beast::system_error{
            beast::error_code{static_cast<int>(::ERR_get_error()), net::error::get_ssl_category()},
            "Failed to set SNI"
        };
    }

    stream.handshake(ssl::stream_base::client);

    http::request<http::string_body> req{http::verb::post, "/api/v1/bullet-public", 11};
    req.set(http::field::host, "api.kucoin.com");
    req.set(http::field::user_agent, "kucoin-connector");

    http::write(stream, req);

    beast::flat_buffer buffer;
    http::response<http::string_body> res;
    http::read(stream, buffer, res);

    beast::error_code ec;
    stream.shutdown(ec);

    auto obj = json::parse(res.body()).as_object();
    return obj["data"].as_object()["token"].as_string().c_str();
}

int main() {
    try {
        net::io_context ioc;
        ssl::context ctx{ssl::context::tlsv12_client};
        ctx.set_default_verify_paths();
        ctx.set_verify_mode(ssl::verify_none);

        std::string token = get_kucoin_token(ioc, ctx);
        std::cout << "Got KuCoin token" << std::endl;

        tcp::resolver resolver{ioc};
        websocket::stream<ssl::stream<tcp::socket>> ws{ioc, ctx};

        auto const results = resolver.resolve("ws-api-spot.kucoin.com", "443");
        auto ep = net::connect(beast::get_lowest_layer(ws), results);

        if(!SSL_set_tlsext_host_name(ws.next_layer().native_handle(), "ws-api-spot.kucoin.com")) {
            throw beast::system_error{
                beast::error_code{static_cast<int>(::ERR_get_error()), net::error::get_ssl_category()},
                "Failed to set SNI"
            };
        }

        ws.next_layer().handshake(ssl::stream_base::client);

        std::string ws_path = "/?token=" + token;
        ws.handshake("ws-api-spot.kucoin.com", ws_path);

        std::string subscribe_msg = R"({"id":"1","type":"subscribe","topic":"/market/candles:BTC-USDT_1min","response":true})";
        ws.write(net::buffer(subscribe_msg));

        std::cout << "Connected to KuCoin WebSocket (BTC-USDT 1m candles)" << std::endl;

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
