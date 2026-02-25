import unittest

from app import build_output


class InferencePipelineTest(unittest.TestCase):
    def test_inference_output_shape(self):
        out = build_output({'symbol': 'AAPL', 'payload': {'log_return': 0.35}})
        self.assertIn('symbol', out)
        self.assertIn('score', out)
        self.assertIn('severity', out)
        self.assertIn('explanation', out)

    def test_malformed_input_fails_safely(self):
        with self.assertRaises(ValueError):
            build_output({'payload': {'log_return': 0.2}})


if __name__ == '__main__':
    unittest.main()
