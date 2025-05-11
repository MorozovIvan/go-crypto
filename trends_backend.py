from flask import Flask, request, jsonify
from flask_cors import CORS
from pytrends.request import TrendReq

app = Flask(__name__)
CORS(app)

@app.route('/api/trends', methods=['GET'])
def get_trends():
    keyword = request.args.get('keyword', 'bitcoin')
    timeframe = request.args.get('timeframe', 'now 7-d')
    pytrends = TrendReq(hl='en-US', tz=360)
    pytrends.build_payload([keyword], cat=0, timeframe=timeframe, geo='', gprop='')
    data = pytrends.interest_over_time()
    if not data.empty:
        values = data[keyword].tolist()
        return jsonify({'values': values, 'labels': list(data.index.strftime('%Y-%m-%d'))})
    else:
        return jsonify({'values': [], 'labels': []})

if __name__ == '__main__':
    app.run(port=5001) 