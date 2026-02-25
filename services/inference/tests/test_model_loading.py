import json
import tempfile
import unittest
from unittest.mock import patch

from app import load_model


class ModelLoadingTest(unittest.TestCase):
    def test_default_model_loaded_without_file(self):
        with patch('os.getenv', return_value=''):
            model = load_model()
            self.assertEqual(model['name'], 'baseline')

    def test_model_loading_failure_is_explicit(self):
        with patch('os.getenv', return_value='/does/not/exist.json'):
            with self.assertRaises(FileNotFoundError):
                load_model()

    def test_model_loading_reads_json(self):
        with tempfile.NamedTemporaryFile(mode='w+', suffix='.json') as f:
            json.dump({'name': 'xgb', 'version': 'v2'}, f)
            f.flush()
            with patch('os.getenv', return_value=f.name):
                model = load_model()
                self.assertEqual(model['version'], 'v2')


if __name__ == '__main__':
    unittest.main()
